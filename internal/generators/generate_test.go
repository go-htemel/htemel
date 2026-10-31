package generators

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/go-htemel/spec"
)

func TestGoIdent(t *testing.T) {
	for in, want := range map[string]string{
		"a":                   "A",
		"div":                 "Div",
		"accept-charset":      "AcceptCharset",
		"http-equiv":          "HttpEquiv",
		"clipPath":            "ClipPath",
		"feGaussianBlur":      "FeGaussianBlur",
		"viewBox":             "ViewBox",
		"preserveAspectRatio": "PreserveAspectRatio",
		"stroke-width":        "StrokeWidth",
		"stroke-linecap":      "StrokeLinecap",
		"xlink:href":          "XlinkHref",
		"xml:space":           "XmlSpace",
		"xmlns:xlink":         "XmlnsXlink",
		"x-ua-compatible":     "XUaCompatible",
		"1":                   "V1",
	} {
		if got := goIdent(in); got != want {
			t.Errorf("goIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCheckNamesCollisions(t *testing.T) {
	str := func(n string) spec.Attribute { return &spec.AttributeTypeString{Name: n} }
	for name, elems := range map[string][]*spec.Element{
		"case-only tags":      {{Tag: "clipPath"}, {Tag: "clippath"}},
		"mangled attributes":  {{Tag: "path", Attributes: []spec.Attribute{str("stroke-width"), str("strokeWidth")}}},
		"built-in method":     {{Tag: "g", Attributes: []spec.Attribute{str("render")}}},
		"void built-in still": {{Tag: "br", Void: true, Attributes: []spec.Attribute{str("attr")}}},
		"SetClass built-in":   {{Tag: "div", Attributes: []spec.Attribute{str("set-class")}}},
	} {
		if err := CheckNames(elems); err == nil {
			t.Errorf("%s: expected a collision error", name)
		}
	}
	if err := CheckNames([]*spec.Element{{Tag: "br", Void: true, Attributes: []spec.Attribute{str("text")}}}); err != nil {
		t.Errorf("void element has no Text method, so a text attribute is fine: %v", err)
	}
	// The retired per-element conditionals no longer reserve names.
	if err := CheckNames([]*spec.Element{{Tag: "div", Attributes: []spec.Attribute{str("if"), str("if-else"), str("if-then")}}, {Tag: "divIf"}}); err != nil {
		t.Errorf("If/IfElse/IfThen/XxxIf are no longer generated: %v", err)
	}
	// class is handled by the built-in Class methods, not a generated setter.
	if err := CheckNames([]*spec.Element{{Tag: "div", Attributes: []spec.Attribute{&spec.AttributeTypeSST{Name: "class"}}}}); err != nil {
		t.Errorf("class attribute must not collide with the built-in Class: %v", err)
	}
}

// TestMoveGlobals checks ADR 0010: only the merged global attributes not in the direct list move
// into Globals; an element's own attribute of the same name keeps its setter; the Globals
// method and struct are checked for collisions.
func TestMoveGlobals(t *testing.T) {
	slot := &spec.AttributeTypeString{Name: "slot"}
	id := &spec.AttributeTypeString{Name: "id"}
	ownSlot := &spec.AttributeTypeString{Name: "slot"}
	sp := &spec.Spec{Name: "HTML", Attributes: []spec.Attribute{id, slot}, Elements: []*spec.Element{
		{Tag: "div"},
		{Tag: "custom", Attributes: []spec.Attribute{ownSlot}},
	}}
	mergeGlobals(sp)
	p, err := NewPlan(sp, "html")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.moveGlobals(sp.Attributes, HTMLDirectGlobals); err != nil {
		t.Fatal(err)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if len(p.Globals) != 1 || p.Globals[0] != spec.Attribute(slot) {
		t.Fatalf("Globals = %v, want [slot]", p.Globals)
	}
	names := func(tag string) (out []string) {
		for _, a := range p.element(tag).Attributes {
			out = append(out, a.GetName())
		}
		return out
	}
	if got := names("div"); !slices.Equal(got, []string{"id"}) {
		t.Errorf("div attributes = %v, want [id]", got)
	}
	if got := names("custom"); !slices.Equal(got, []string{"slot", "id"}) || p.element("custom").Attributes[0] != spec.Attribute(ownSlot) {
		t.Errorf("custom must keep its own slot setter: %v", got)
	}
	if !p.element("div").HasGlobals {
		t.Error("div must get the Globals method")
	}

	// A global named "globals" would collide with the Globals method.
	clash := &spec.Spec{Name: "HTML", Attributes: []spec.Attribute{slot}, Elements: []*spec.Element{
		{Tag: "div", Attributes: []spec.Attribute{&spec.AttributeTypeString{Name: "globals"}}},
	}}
	mergeGlobals(clash)
	p, err = NewPlan(clash, "html")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.moveGlobals(clash.Attributes, HTMLDirectGlobals); err != nil {
		t.Fatal(err)
	}
	if err := p.Check(); err == nil || !strings.Contains(err.Error(), "Globals") {
		t.Errorf("expected a Globals collision, got %v", err)
	}

	// Prefixed families cannot become struct fields.
	fam := &spec.Spec{Name: "HTML", Attributes: []spec.Attribute{&spec.AttributeTypePrefixedCustom{Name: "x"}}, Elements: []*spec.Element{{Tag: "div"}}}
	mergeGlobals(fam)
	if p, err = NewPlan(fam, "html"); err != nil {
		t.Fatal(err)
	}
	if err := p.moveGlobals(fam.Attributes, HTMLDirectGlobals); err == nil {
		t.Error("expected an error for a prefixed global family")
	}
}

// TestGenerateSVGFixture generates a small SVG spec, parses every file, checks the case-sensitive and
// mangled names, then builds the package and renders a tree with it.
func TestGenerateSVGFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "svg_fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	sp := spec.Spec{}
	if err = json.Unmarshal(b, &sp); err != nil {
		t.Fatal(err)
	}

	// Generate inside the module (under testdata, so ./... never sees it) so the package can import
	// github.com/go-htemel/htemel/internal/render.
	dir, err := os.MkdirTemp("testdata", "gen-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	pkgDir := filepath.Join(dir, "svg")

	if err = Generate(&sp, pkgDir, "svg"); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	all := &strings.Builder{}
	for _, name := range []string{"enums.go", "content.go"} {
		src, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			t.Fatal(err)
		}
		all.Write(src)
	}
	for _, e := range sp.Elements {
		path := filepath.Join(pkgDir, strings.ToLower(e.Tag)+".go")
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected lower-cased file for <%s>: %v", e.Tag, err)
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name.Name != "svg" {
			t.Errorf("%s: package %s, want svg", path, f.Name.Name)
		}
		all.Write(src)
	}
	code := strings.Join(strings.Fields(all.String()), " ") // whitespace-insensitive (gofmt aligns)

	for _, want := range []string{
		"func ClipPath(",
		"type ClipPathElement struct",
		"func LinearGradient(",
		"func FeGaussianBlur(",
		"func (e *PathElement) StrokeWidth(s string) *PathElement",
		"func (e *PathElement) StrokeLinecap(a StrokeLinecap) *PathElement",
		"StrokeLinecapButt StrokeLinecap = \"butt\"",
		"func Svg(children ...Content) *SvgElement",
		"htemel.PhrasingMarker",
		"func (e *SvgElement) ViewBox(",
		"func (e *SvgElement) PreserveAspectRatio(",
		"func (e *SvgElement) XmlnsXlink(",
		"func (e *LinearGradientElement) XlinkHref(",
		"func (e *GElement) XmlSpace(",
		"func (e *CircleElement) Opacity(f float64)",
		"func (e *CircleElement) Tabindex(i int)",
		"func (e *RectElement) Class(classes ...string)",
		"func (e *RectElement) SetClass(classes ...string)",
		`"stroke-width"`,
		`"stroke-linecap"`,
		`"viewBox"`,
		`"preserveAspectRatio"`,
		`"xlink:href"`,
		`"xml:space"`,
		`"<clipPath"`,
		`"</clipPath>"`,
		`"<feGaussianBlur"`,
		`"<linearGradient"`,
		`render.RawText("script", text)`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("generated code is missing %s", want)
		}
	}
	for _, bad := range []string{"Clippath", "Lineargradient", "Fegaussianblur", "Viewbox", "<clippath", `"viewbox"`} {
		if strings.Contains(code, bad) {
			t.Errorf("generated code contains case-folded name %s", bad)
		}
	}

	if testing.Short() {
		return
	}

	mainDir := filepath.Join(dir, "main")
	if err = os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const mainSrc = `package main

import (
	"fmt"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/internal/generators/` + "DIR" + `/svg"
)

func main() {
	out, err := htemel.RenderString(
		svg.Svg(
			svg.ClipPath().Id("c"),
			svg.Path().D("M0 0L1 1").Stroke("currentColor").StrokeWidth("2").StrokeLinecap(svg.StrokeLinecapRound),
			svg.FeGaussianBlur().StdDeviation("3"),
			svg.Script().Text("if (a < b) {}"),
		).ViewBox("0 0 24 24").XmlSpace("preserve"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
`
	src := strings.Replace(mainSrc, "DIR", filepath.ToSlash(dir), 1)
	if err = os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./"+filepath.ToSlash(mainDir))
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building/running the generated package failed: %v\n%s", err, got)
	}
	want := `<svg viewBox="0 0 24 24" xml:space="preserve"><clipPath id="c"></clipPath>` +
		`<path d="M0 0L1 1" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>` +
		`<feGaussianBlur stdDeviation="3"></feGaussianBlur><script>if (a < b) {}</script></svg>`
	if string(got) != want {
		t.Errorf("rendered\n%s\nwant\n%s", got, want)
	}
}

// TestGenerateSpecFlags checks the doc comments and raw-text handling driven by the spec's
// deprecated, url and raw_text fields.
func TestGenerateSpecFlags(t *testing.T) {
	sp := spec.Spec{
		Name: "Flags",
		Elements: []*spec.Element{
			{Tag: "old", Description: "old", Deprecated: true},
			{Tag: "raw", Description: "raw", RawText: true},
			{Tag: "script", Description: "legacy spec without flag"},
			{Tag: "plain", Description: "plain", Attributes: []spec.Attribute{
				&spec.AttributeTypeString{Name: "href", URL: true},
				&spec.AttributeTypeString{Name: "title"},
				&spec.AttributeTypeNumber{Name: "size", Deprecated: true},
				&spec.AttributeTypeBool{Name: "hidden", Deprecated: true},
			}},
		},
	}
	dir := t.TempDir()
	if err := Generate(&sp, dir, "flags"); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	if old := read("old.go"); !strings.Contains(old, "// Deprecated: <old>") {
		t.Errorf("deprecated element missing Deprecated doc:\n%s", old)
	}
	if plain := read("plain.go"); strings.Contains(plain, "// Deprecated: <plain>") {
		t.Error("non-deprecated element must not carry a Deprecated doc")
	}
	for _, tag := range []string{"raw", "script"} {
		if src := read(tag + ".go"); !strings.Contains(src, "render.RawText(\""+tag+"\"") {
			t.Errorf("<%s> should be a raw text element", tag)
		}
	}
	if old := read("old.go"); strings.Contains(old, "render.RawText") {
		t.Error("<old> must not be raw text")
	}

	plain := read("plain.go")
	for _, want := range []string{
		"// Value is a URL; validate untrusted input with htemel.URL.\nfunc (e *PlainElement) Href(",
		"// Deprecated: the \"size\" attribute",
		"// Deprecated: the \"hidden\" attribute",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("plain.go missing %q:\n%s", want, plain)
		}
	}
	if i := strings.Index(plain, "func (e *PlainElement) Title("); i < 0 || strings.Contains(plain[:i][strings.LastIndex(plain[:i], "}"):], "URL") {
		t.Error("non-URL string attribute must not get the URL doc")
	}
}

func TestCheckNamesContentModelCollisions(t *testing.T) {
	ul := func() *spec.Element {
		return &spec.Element{Tag: "ul", ContentModel: &spec.ContentModel{Kind: spec.ContentKindChildren, Elements: []string{"li"}}}
	}
	li := func(attrs ...spec.Attribute) *spec.Element {
		return &spec.Element{Tag: "li", Attributes: attrs, ContentModel: &spec.ContentModel{Kind: spec.ContentKindChildren, Categories: []string{"flow"}, Text: true}}
	}
	str := func(n string) spec.Attribute { return &spec.AttributeTypeString{Name: n} }
	for name, elems := range map[string][]*spec.Element{
		"element named like a child interface":    {ul(), li(), {Tag: "ulChild"}},
		"attribute named like a marker method":    {ul(), li(str("HTMLUlChild"))},
		"attribute named like an embedded marker": {{Tag: "div", Categories: []string{"flow"}, Attributes: []spec.Attribute{str("FlowMarker")}}},
	} {
		p, err := NewPlan(&spec.Spec{Name: "HTML", Elements: elems}, "html")
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Check(); err == nil {
			t.Errorf("%s: expected a collision error", name)
		}
	}

	// A shared enum name that collides with an element falls back to per-element types.
	pop := &spec.AttributeTypeEnum{Name: "popover", Allowed: map[string]struct{}{"auto": {}}}
	p, err := NewPlan(&spec.Spec{Name: "HTML", Elements: []*spec.Element{
		{Tag: "div", Attributes: []spec.Attribute{pop}},
		{Tag: "popoverAuto"},
	}}, "html")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if _, shared := p.SharedEnums["popover"]; shared || len(p.EnumReport) != 1 || p.EnumType("div", "popover") != "DivPopover" {
		t.Errorf("popover should fall back to DivPopover: shared=%v report=%v", shared, p.EnumReport)
	}
}

// TestGenerateContentModelFixture generates a small HTML spec with content models covering every
// kind, checks the generated signatures and builds a program using the package.
func TestGenerateContentModelFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "html_cm_fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	sp := spec.Spec{}
	if err = json.Unmarshal(b, &sp); err != nil {
		t.Fatal(err)
	}

	dir, err := os.MkdirTemp("testdata", "gen-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	pkgDir := filepath.Join(dir, "html")
	slots := filepath.Join(dir, "slots.go.txt")

	if err = GenerateWith(&sp, pkgDir, "html", Options{SlotsFile: slots, DirectGlobals: HTMLDirectGlobals}); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(strings.Fields(string(b)), " ")
	}
	file := func(tag string) string { return read(filepath.Join(pkgDir, tag+".go")) }

	for tag, wants := range map[string][]string{
		"div":      {"func Div(children ...htemel.FlowContent) *DivElement", "htemel.FlowMarker htemel.PalpableMarker attributes", "func (e *DivElement) Text(", "func (e *DivElement) Globals(g Globals) *DivElement", "func (e *DivElement) Class(classes ...string) *DivElement", "func (e *DivElement) SetClass(classes ...string) *DivElement", "func (e *DivElement) Tabindex(i int) *DivElement", "func (*DivElement) HTMLSelectChild() {}", "func (e *DivElement) Popover(v ...Popover) *DivElement"},
		"p":        {"func P(children ...htemel.PhrasingContent) *PElement"},
		"span":     {"htemel.PhrasingMarker htemel.PalpableMarker attributes", "func (e *SpanElement) Popover(v ...Popover)"},
		"ul":       {"type UlChild interface { htemel.Node HTMLUlChild() }", "func Ul(children ...UlChild) *UlElement", "func (e *UlElement) Type(a UlType)", "type UlType string"},
		"li":       {"func (*LiElement) HTMLUlChild() {}", "func Li(children ...htemel.FlowContent)"},
		"table":    {"func Table(children ...TableChild)"},
		"tr":       {"func Tr(children ...TrChild)", "func (*TrElement) HTMLTableChild() {}"},
		"td":       {"func (*TdElement) HTMLTrChild() {}"},
		"a":        {"func A(children ...htemel.FlowContent)", "htemel.PhrasingMarker htemel.InteractiveMarker htemel.PalpableMarker", "Content model: transparent."},
		"img":      {"func Img() *ImgElement", "func (e *ImgElement) Globals(g Globals) *ImgElement"},
		"iframe":   {"func Iframe() *IframeElement", `"</iframe>"`, "takes no children"},
		"script":   {"func Script(children ...htemel.TextContent)", `render.RawText("script", text)`, "type ScriptType string", "ScriptTypeModule ScriptType = \"module\"", "func ScriptTypeCustom(s string) ScriptType"},
		"title":    {"func Title(children ...htemel.TextContent)"},
		"h1":       {"func H1(children ...htemel.PhrasingContent)", "htemel.HeadingMarker"},
		"head":     {"func Head(children ...htemel.MetadataContent)"},
		"meta":     {"func Meta() *MetaElement", "htemel.PhrasingMarker htemel.MetadataMarker"},
		"html":     {"type HtmlChild interface", "func Html(children ...HtmlChild)"},
		"body":     {"func (*BodyElement) HTMLHtmlChild() {}", "func Body(children ...htemel.FlowContent)"},
		"template": {"func Template(children ...htemel.Node)", "func (e *TemplateElement) Text("},
		"select":   {"func Select(children ...SelectChild)"},
		"option":   {"func Option(children ...OptionChild)", "func (*OptionElement) HTMLSelectChild() {}", "func (e *OptionElement) Text("},
	} {
		src := file(tag)
		for _, want := range wants {
			if !strings.Contains(src, want) {
				t.Errorf("%s.go is missing %s", tag, want)
			}
		}
	}
	for tag, bads := range map[string][]string{
		"img":    {"Children(", "Text(", "children []"},
		"iframe": {"Children(", "Text("},
		"div":    {"DivIf", "DivTernary", ") If(", ") IfElse(", ") IfThen(", "skipRender", ") Itemscope(", ") Itemprop(", ") Translate("},
		"meta":   {"Children("},
		"ul":     {"func (e *UlElement) Text("},
		"script": {"HTMLUlChild"}, // promoted from htemel.ScriptSupportingMarker
		"span":   {"HTMLOptionChild"},
	} {
		src := file(tag)
		for _, bad := range bads {
			if strings.Contains(src, bad) {
				t.Errorf("%s.go must not contain %s", tag, bad)
			}
		}
	}

	enums := read(filepath.Join(pkgDir, "enums.go"))
	for _, want := range []string{"type Popover string", "PopoverManual Popover = \"manual\"", "PopoverEmpty Popover = \"\"", "- type: allowed values differ between elements (Script, Ul)"} {
		if !strings.Contains(enums, want) {
			t.Errorf("enums.go is missing %s", want)
		}
	}
	globalsSrc := read(filepath.Join(pkgDir, "globals.go"))
	for _, want := range []string{"type Globals struct {", "Itemscope bool", "Itemprop []string", "Translate Translate", "func (g Globals) apply(a *render.Attrs)"} {
		if !strings.Contains(globalsSrc, want) {
			t.Errorf("globals.go is missing %s", want)
		}
	}
	for _, bad := range []string{"Popover", "Class", "Tabindex"} {
		if strings.Contains(globalsSrc, bad) {
			t.Errorf("globals.go must not contain the direct global %s", bad)
		}
	}
	if i, j := strings.Index(globalsSrc, `"itemscope"`), strings.Index(globalsSrc, `"translate"`); i < 0 || i > j {
		t.Error("Globals must apply fields in spec order")
	}

	slotSrc := read(slots)
	for _, want := range []string{"package htemel", "func (Wildcard) HTMLUlChild() {}", "func (NodeFunc) HTMLUlChild() {}", "func (ScriptSupportingMarker) HTMLUlChild() {}", "func (*TextElement) HTMLOptionChild() {}", "func (PhrasingMarker) HTMLOptionChild() {}"} {
		if !strings.Contains(slotSrc, want) {
			t.Errorf("slots file is missing %s", want)
		}
	}
	if strings.Contains(slotSrc, "func (*TextElement) HTMLUlChild()") {
		t.Error("text must not be a permitted child of <ul>")
	}

	if testing.Short() {
		return
	}

	// The package builds and renders. (The marker methods of the real html package, in package
	// htemel, have the same names, so Group and Each are accepted here too.)
	mainDir := filepath.Join(dir, "main")
	if err = os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := strings.Replace(`package main

import (
	"fmt"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/internal/generators/DIR/html"
)

func main() {
	out, err := htemel.RenderString(html.Html(
		html.Head(html.Title().Text("t"), html.Meta(), html.Script().Text("a < b")),
		html.Body(
			html.Div(html.P(html.Span().Text("x & y"), html.A(html.Img())), html.H1().Text("h")),
			html.Ul(html.Li().Text("1"), htemel.Each([]string{"2"}, func(s string) htemel.Node { return html.Li().Text(s) }), html.Script()),
			html.Table(html.Tr(html.Td().Text("c"))),
			html.Iframe().Popover(html.PopoverManual),
			html.Div().Class("a b").Class("b c").Globals(html.Globals{Translate: html.TranslateNo, Itemscope: true, Itemprop: []string{"x", "y"}}).Tabindex(1),
			html.Select(html.Option(html.Span().Text("o")), html.Div()),
			html.Template(html.Tr(), html.Li()),
		),
	))
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
`, "DIR", filepath.ToSlash(dir), 1)
	if err = os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := exec.Command("go", "run", "./"+filepath.ToSlash(mainDir)).CombinedOutput()
	if err != nil {
		t.Fatalf("building/running the generated package failed: %v\n%s", err, got)
	}
	want := `<html><head><title>t</title><meta><script>a < b</script></head><body>` +
		`<div><p><span>x &amp; y</span><a><img></a></p><h1>h</h1></div>` +
		`<ul><li>1</li><li>2</li><script></script></ul><table><tr><td>c</td></tr></table>` +
		`<iframe popover="manual"></iframe><div class="a b c" itemscope itemprop="x y" translate="no" tabindex="1"></div><select><option><span>o</span></option><div></div></select>` +
		`<template><tr></tr><li></li></template></body></html>`
	if string(got) != want {
		t.Errorf("rendered\n%s\nwant\n%s", got, want)
	}
}
