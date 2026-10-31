package generators

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/go-htemel/spec"
)

// categoryNames maps each content category that has an interface in package htemel to the base
// name of that interface (htemel.<Name>Content) and of its marker struct (htemel.<Name>Marker).
// Other categories (form-associated, listed, ...) have no interface; they still take part in
// child-interface membership, where an element of the category gets the method explicitly.
var categoryNames = map[string]string{
	"flow":              "Flow",
	"phrasing":          "Phrasing",
	"metadata":          "Metadata",
	"sectioning":        "Sectioning",
	"heading":           "Heading",
	"embedded":          "Embedded",
	"interactive":       "Interactive",
	"palpable":          "Palpable",
	"script-supporting": "ScriptSupporting",
}

// categoryOrder fixes the order of markers and categories in generated code.
var categoryOrder = []string{"flow", "phrasing", "metadata", "sectioning", "heading", "embedded", "interactive", "palpable", "script-supporting"}

// foreignRoots maps the (lower-cased) name of a spec without HTML content models to the tag of its
// root element. These packages get a closed, package-local model: every element accepts
// ...<pkg>.Content, which every element of the package (and text) implements. The root element is
// also embedded, phrasing, flow and palpable content, as the HTML specification categorizes
// <svg> and <math>, so it can be placed inside HTML elements.
var foreignRoots = map[string]string{"svg": "svg", "mathml": "math"}

var foreignRootCategories = []string{"embedded", "flow", "palpable", "phrasing"}

// foreignUnchecked lists elements of foreign packages that host HTML content (and so accept any
// htemel.Node as a child).
var foreignUnchecked = map[string]map[string]bool{
	"svg":    {"foreignObject": true},
	"mathml": {"annotation-xml": true},
}

// methodPrefixes gives the prefix of child-interface method names per package, so that the
// methods of different packages never clash on htemel.Wildcard (HTMLUlChild, SVGContent).
var methodPrefixes = map[string]string{"html": "HTML", "svg": "SVG", "mathml": "MathML"}

// Generated file names that are not element files.
const (
	enumsFile   = "enums.go"
	contentFile = "content.go"
)

// Plan is everything the generator decides about a package before writing it: the child type
// of every element, the element-specific child interfaces and the enum types.
type Plan struct {
	Pkg     string
	Prefix  string
	Foreign bool

	Elements []*ElementPlan
	byTag    map[string]*ElementPlan

	// Slots are the generated child interfaces (html.UlChild, svg.Content), sorted by name.
	Slots []*Slot

	// SharedEnums are the enum types shared by every element with the attribute, by attribute name.
	SharedEnums map[string]*SharedEnum
	// EnumReport lists enum attributes that could not be shared, and why.
	EnumReport []string
	// Globals are the global attributes moved into the generated Globals struct (ADR 0010), in
	// spec order; empty when every global is a direct method.
	Globals []spec.Attribute

	cats map[string]map[string]bool // tag -> effective categories
}

// ElementPlan is the plan for one element.
type ElementPlan struct {
	*spec.Element

	// Markers are the htemel marker structs embedded in the element struct.
	Markers []string
	// HasChildren reports whether the element takes children (constructor parameter and
	// Children).
	HasChildren bool
	// ChildType is the Go type of a child, e.g. htemel.FlowContent or UlChild.
	ChildType string
	// Text reports whether the element has Text and Textf methods.
	Text bool
	// HasGlobals reports whether the element has the Globals method.
	HasGlobals bool
	// Slot is the child interface declared in this element's file, if any.
	Slot *Slot
	// SlotMethods are the child interfaces whose marker method is generated on this element;
	// methods promoted from its markers are not repeated.
	SlotMethods []*Slot
	// memberOf lists every child interface the element implements.
	memberOf []*Slot
}

// Slot is a generated child interface: the permitted children of an element whose content model
// is not a single category.
type Slot struct {
	Owner  string // tag declaring it; "" for a package-wide interface (foreign packages)
	Iface  string // interface name, e.g. UlChild
	Method string // marker method, e.g. HTMLUlChild
	// Categories are the permitted categories, reduced so that none is implied by another.
	Categories []string
	// Elements are the explicitly permitted tags not already covered by Categories.
	Elements []string
	// Text reports whether text is permitted.
	Text bool
	// Members are the tags of this package implementing the interface.
	Members []string
}

// Carriers are the htemel marker structs that carry the slot method, so that any type embedding
// one of them (including elements of other packages, such as svg.Svg) implements the interface.
func (s *Slot) Carriers() []string {
	var out []string
	for _, c := range s.Categories {
		if n, ok := categoryNames[c]; ok {
			out = append(out, n+"Marker")
		}
	}
	return out
}

// SharedEnum is an enum type named after its attribute and shared by every element.
type SharedEnum struct {
	Type string
	Attr *spec.AttributeTypeEnum
	Tags []string
}

func (p *Plan) element(tag string) *ElementPlan { return p.byTag[tag] }

// EnumType returns the Go type of the enum attribute name on tag.
func (p *Plan) EnumType(tag, name string) string {
	if s, ok := p.SharedEnums[name]; ok {
		return s.Type
	}
	return goIdent(tag) + goIdent(name)
}

// NewPlan plans the package pkg for sp, whose global attributes must already be merged into its
// elements.
func NewPlan(sp *spec.Spec, pkg string) (*Plan, error) {
	p := &Plan{
		Pkg:         pkg,
		Prefix:      methodPrefixes[pkg],
		byTag:       map[string]*ElementPlan{},
		SharedEnums: map[string]*SharedEnum{},
		cats:        map[string]map[string]bool{},
	}
	if p.Prefix == "" {
		p.Prefix = goIdent(pkg)
	}
	root, foreign := foreignRoots[strings.ToLower(sp.Name)]
	p.Foreign = foreign

	for _, e := range sp.Elements {
		ep := &ElementPlan{Element: e}
		p.Elements = append(p.Elements, ep)
		p.byTag[e.Tag] = ep

		cats := map[string]bool{}
		for _, c := range e.Categories {
			cats[c] = true
		}
		if foreign && e.Tag == root && len(cats) == 0 {
			for _, c := range foreignRootCategories {
				cats[c] = true
			}
		}
		if cats["phrasing"] {
			cats["flow"] = true
		}
		p.cats[e.Tag] = cats

		for _, c := range categoryOrder {
			if !cats[c] || (c == "flow" && cats["phrasing"]) {
				continue
			}
			ep.Markers = append(ep.Markers, "htemel."+categoryNames[c]+"Marker")
		}
	}

	var content *Slot
	if foreign {
		content = &Slot{Iface: "Content", Method: p.Prefix + "Content", Text: true}
		for _, ep := range p.Elements {
			content.Members = append(content.Members, ep.Tag)
		}
		p.Slots = append(p.Slots, content)
	}

	for _, ep := range p.Elements {
		switch {
		case foreign && ep.Void:
		case foreign && foreignUnchecked[strings.ToLower(sp.Name)][ep.Tag]:
			ep.HasChildren, ep.ChildType, ep.Text = true, "htemel.Node", true
		case foreign:
			ep.HasChildren, ep.ChildType, ep.Text = true, "Content", true
		default:
			p.resolveHTML(ep)
		}
	}

	for _, s := range p.Slots {
		if s.Owner != "" {
			for _, ep := range p.Elements {
				if p.isMember(ep.Tag, s) {
					s.Members = append(s.Members, ep.Tag)
				}
			}
		}
		for _, tag := range s.Members {
			ep := p.element(tag)
			ep.memberOf = append(ep.memberOf, s)
			if !p.promoted(ep, s) {
				ep.SlotMethods = append(ep.SlotMethods, s)
			}
		}
	}
	sort.Slice(p.Slots, func(i, j int) bool { return p.Slots[i].Iface < p.Slots[j].Iface })

	p.planEnums()
	return p, nil
}

// resolveHTML sets the child type of an element from its content model. Elements without a
// content model (old spec files) accept any htemel.Node unless they are void.
func (p *Plan) resolveHTML(ep *ElementPlan) {
	m := ep.ContentModel
	switch {
	case m == nil:
		if !ep.Void {
			ep.HasChildren, ep.ChildType, ep.Text = true, "htemel.Node", true
		}
	case m.Kind == spec.ContentKindEmpty:
	case m.Kind == spec.ContentKindText:
		ep.HasChildren, ep.ChildType, ep.Text = true, "htemel.TextContent", true
	case m.Kind == spec.ContentKindTransparent:
		// The element takes its parent's model, which a type cannot express: accept flow content
		// (the widest model a transparent element can inherit in a body) plus the listed tags.
		p.resolveChildren(ep, []string{"flow"}, m.Elements, true)
	case m.Kind == spec.ContentKindChildren:
		p.resolveChildren(ep, m.Categories, m.Elements, m.Text)
	default: // unknown (template) or a kind this generator does not know: unchecked
		ep.HasChildren, ep.ChildType, ep.Text = true, "htemel.Node", true
	}
}

func (p *Plan) resolveChildren(ep *ElementPlan, categories, elements []string, text bool) {
	cats := p.reduce(categories)
	hasTextCat := slices.Contains(cats, "flow") || slices.Contains(cats, "phrasing")

	var tags []string
	for _, tag := range elements {
		if _, ok := p.byTag[tag]; !ok {
			continue // not an element of this package
		}
		covered := false
		for _, c := range cats {
			if p.cats[tag][c] {
				covered = true
			}
		}
		if !covered {
			tags = append(tags, tag)
		}
	}
	explicitText := text && !hasTextCat

	ep.Text = text || hasTextCat
	switch {
	case len(tags) == 0 && len(cats) == 0 && text:
		ep.HasChildren, ep.ChildType = true, "htemel.TextContent"
		return
	case len(tags) == 0 && len(cats) == 0:
		return // nothing permitted
	case len(tags) == 0 && len(cats) == 1 && !explicitText && categoryNames[cats[0]] != "":
		ep.HasChildren, ep.ChildType = true, "htemel."+categoryNames[cats[0]]+"Content"
		return
	}

	s := &Slot{
		Owner:      ep.Tag,
		Iface:      goIdent(ep.Tag) + "Child",
		Method:     p.Prefix + goIdent(ep.Tag) + "Child",
		Categories: cats,
		Elements:   tags,
		Text:       ep.Text,
	}
	p.Slots = append(p.Slots, s)
	ep.Slot = s
	ep.HasChildren, ep.ChildType = true, s.Iface
}

// subsumed reports whether every element of category c is also in category d (and c is not
// empty), judged from this package's elements.
func (p *Plan) subsumed(c, d string) bool {
	found := false
	for _, cats := range p.cats {
		if cats[c] {
			found = true
			if !cats[d] {
				return false
			}
		}
	}
	return found
}

// reduce orders categories and drops those implied by another listed category (phrasing when
// flow is listed, script-supporting when phrasing is listed, ...). Besides keeping the generated
// types small, this guarantees that no element embeds two markers carrying the same slot
// method, which would make the method ambiguous and silently drop it.
func (p *Plan) reduce(categories []string) []string {
	in := slices.Clone(categories)
	rank := func(c string) int {
		if i := slices.Index(categoryOrder, c); i >= 0 {
			return i
		}
		return len(categoryOrder)
	}
	sort.SliceStable(in, func(i, j int) bool {
		if rank(in[i]) != rank(in[j]) {
			return rank(in[i]) < rank(in[j])
		}
		return in[i] < in[j]
	})
	in = slices.Compact(in)
	var out []string
	for i, c := range in {
		drop := false
		for j, d := range in {
			if i == j || !p.subsumed(c, d) {
				continue
			}
			if p.subsumed(d, c) && j > i {
				continue // same set: keep the first
			}
			drop = true
		}
		if !drop {
			out = append(out, c)
		}
	}
	return out
}

func (p *Plan) isMember(tag string, s *Slot) bool {
	if slices.Contains(s.Elements, tag) {
		return true
	}
	for _, c := range s.Categories {
		if p.cats[tag][c] {
			return true
		}
	}
	return false
}

// promoted reports whether ep gets s.Method unambiguously from its embedded markers.
func (p *Plan) promoted(ep *ElementPlan, s *Slot) bool {
	carriers := s.Carriers()
	hits := 0
	for _, m := range ep.Markers {
		if slices.Contains(carriers, strings.TrimPrefix(m, "htemel.")) {
			hits++
		}
	}
	switch hits {
	case 1:
		return true
	case 0:
		// FlowMarker is embedded at depth 2 through PhrasingMarker.
		return slices.Contains(ep.Markers, "htemel.PhrasingMarker") && slices.Contains(carriers, "FlowMarker")
	}
	return false
}

// planEnums decides which enum attributes get one shared type named after the attribute.
func (p *Plan) planEnums() {
	type use struct {
		tag  string
		attr *spec.AttributeTypeEnum
	}
	uses := map[string][]use{}
	for _, ep := range p.Elements {
		for _, a := range ep.Attributes {
			if en, ok := a.(*spec.AttributeTypeEnum); ok {
				uses[en.Name] = append(uses[en.Name], use{ep.Tag, en})
			}
		}
	}

	taken := map[string]string{}
	for name, what := range p.baseNames() {
		taken[name] = what
	}

	for _, name := range slices.Sorted(maps.Keys(uses)) {
		us := uses[name]
		first := us[0].attr
		var differ []string
		for _, u := range us[1:] {
			if !sameEnum(first, u.attr) {
				differ = append(differ, goIdent(u.tag))
			}
		}
		tags := make([]string, len(us))
		for i, u := range us {
			tags[i] = goIdent(u.tag)
		}
		if len(differ) > 0 {
			p.EnumReport = append(p.EnumReport, fmt.Sprintf("%s: allowed values differ between elements (%s)", name, strings.Join(tags, ", ")))
			continue
		}

		typ := goIdent(name)
		names := enumNames(typ, first)
		var clash []string
		for _, n := range names {
			if what, ok := taken[n]; ok {
				clash = append(clash, n+" ("+what+")")
			}
		}
		if len(clash) > 0 {
			p.EnumReport = append(p.EnumReport, fmt.Sprintf("%s: shared names would collide with %s", name, strings.Join(clash, ", ")))
			continue
		}
		for _, n := range names {
			taken[n] = fmt.Sprintf("enum attribute %q", name)
		}
		p.SharedEnums[name] = &SharedEnum{Type: typ, Attr: first, Tags: tags}
	}
}

func sameEnum(a, b *spec.AttributeTypeEnum) bool {
	return a.AllowEmpty == b.AllowEmpty && a.AllowCustom == b.AllowCustom &&
		maps.Equal(a.Allowed, b.Allowed)
}

// enumNames lists the package-level identifiers an enum type named typ declares.
func enumNames(typ string, a *spec.AttributeTypeEnum) []string {
	names := []string{typ}
	for _, v := range slices.Sorted(maps.Keys(a.Allowed)) {
		names = append(names, typ+goIdent(v))
	}
	if a.AllowEmpty {
		names = append(names, typ+"Empty")
	}
	if a.AllowCustom {
		names = append(names, typ+"Custom")
	}
	return names
}

// baseNames are the package-level identifiers declared regardless of enums.
func (p *Plan) baseNames() map[string]string {
	out := map[string]string{}
	for _, ep := range p.Elements {
		where := "<" + ep.Tag + ">"
		tag := goIdent(ep.Tag)
		out[tag] = where
		out[tag+"Element"] = where
	}
	for _, s := range p.Slots {
		out[s.Iface] = "child interface " + s.Iface
	}
	return out
}

// categoryPhrase describes a list of categories, e.g. "flow content" or "script-supporting elements".
func categoryPhrase(cats []string) []string {
	var out []string
	for _, c := range cats {
		switch c {
		case "script-supporting":
			out = append(out, "script-supporting elements")
		default:
			out = append(out, c+" content")
		}
	}
	return out
}

// Describe lists what the slot permits, e.g. "<li>, script-supporting elements".
func (s *Slot) Describe() string {
	var parts []string
	for _, t := range s.Elements {
		parts = append(parts, "<"+t+">")
	}
	parts = append(parts, categoryPhrase(s.Categories)...)
	if s.Text && !slices.Contains(s.Categories, "flow") && !slices.Contains(s.Categories, "phrasing") {
		parts = append(parts, "text")
	}
	return strings.Join(parts, ", ")
}

// Doc is the doc comment of the slot interface.
func (s *Slot) Doc() string {
	text := "// " + wrapComment(s.Iface+" is a permitted child of <"+s.Owner+">: "+s.Describe()+".") + "\n//\n"
	impl := "It is implemented by " + strings.Join(wrapTags(s.Members), ", ")
	if s.Text {
		impl += ", by text (htemel.Text, htemel.TextUnsafe)"
	}
	if c := s.Carriers(); len(c) > 0 {
		impl += ", by any type embedding htemel." + strings.Join(c, " or htemel.")
	}
	impl += " and by the unchecked nodes of package htemel (Group, Each, If, Unchecked, ...)."
	return text + "// " + wrapComment(impl)
}

// ModelDoc is the doc-comment paragraph describing the element's permitted children.
func (e *ElementPlan) ModelDoc() string {
	m := e.ContentModel
	var line string
	switch {
	case e.ChildType == "Content":
		line = "Children: any element of this package, or text (Content)."
	case e.Void:
		return ""
	case m == nil && e.ChildType == "htemel.Node":
		return ""
	case !e.HasChildren:
		line = "<" + e.Tag + "> takes no children (its content model is nothing), but it is not a void element: it renders an end tag."
	case e.ChildType == "htemel.Node":
		line = "Children are not checked (htemel.Node): the content model cannot be expressed as a type."
	case m != nil && m.Kind == spec.ContentKindTransparent:
		line = "Content model: transparent. <" + e.Tag + "> takes the content model of its parent, which a type cannot follow, so it accepts any flow content"
		if e.Slot != nil && len(e.Slot.Elements) > 0 {
			line += " plus " + strings.Join(wrapTags(e.Slot.Elements), ", ") + " (" + e.ChildType + ")"
		} else {
			line += " (" + e.ChildType + ")"
		}
		line += "; make sure its content is also valid in the parent."
	case e.ChildType == "htemel.TextContent":
		line = "Children: text only (htemel.TextContent)."
	case e.Slot != nil:
		line = "Children: " + e.Slot.Describe() + " (" + e.ChildType + ")."
	default:
		c := strings.TrimSuffix(strings.TrimPrefix(e.ChildType, "htemel."), "Content")
		for k, v := range categoryNames {
			if v == c {
				line = "Children: " + categoryPhrase([]string{k})[0] + " (" + e.ChildType + ")."
			}
		}
	}
	if line == "" {
		return ""
	}
	return "//\n// " + wrapComment(line) + "\n"
}

func wrapTags(tags []string) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = "<" + t + ">"
	}
	return out
}

// wrapComment wraps text at about 100 columns as continuation lines of a // comment.
func wrapComment(text string) string {
	var b strings.Builder
	n := 0
	for i, w := range strings.Fields(text) {
		if i > 0 {
			if n+1+len(w) > 96 {
				b.WriteString("\n// ")
				n = 0
			} else {
				b.WriteByte(' ')
				n++
			}
		}
		b.WriteString(w)
		n += len(w)
	}
	return b.String()
}
