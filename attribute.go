package htemel

// Attribute is a single attribute as a value, so that helper packages (such as
// github.com/go-htemel/htemel/datastar) can build attributes without knowing the concrete element
// type they will be applied to. Apply it with the generated Attrs method that every element has:
//
//	html.Button().Attrs(htemel.Attr("data-on:click", "@post('/save')"))
//
// Name is written as-is and must be a valid attribute name. Value is HTML-escaped when rendered
// unless Unsafe is set. An empty Value renders a bare attribute (name only).
//
// The zero Attribute (empty Name) is a no-op when applied, which makes conditional attributes easy;
// see Attribute.If.
type Attribute struct {
	Name   string
	Value  string
	Unsafe bool
}

// Attr returns an Attribute whose value is HTML-escaped when rendered.
func Attr(name, value string) Attribute {
	return Attribute{Name: name, Value: value}
}

// AttrUnsafe returns an Attribute whose value is rendered verbatim, without escaping.
// Never pass untrusted input.
func AttrUnsafe(name, value string) Attribute {
	return Attribute{Name: name, Value: value, Unsafe: true}
}

// If returns a unless cond is false, in which case it returns the zero Attribute, which is
// ignored when applied.
func (a Attribute) If(cond bool) Attribute {
	if cond {
		return a
	}
	return Attribute{}
}

// Element is implemented by every generated element (in the html and svg packages). It is the hook
// third-party packages use to set attributes without depending on a concrete element type.
//
// Setting an existing attribute updates its value but keeps its original position.
type Element interface {
	Node
	// SetAttr sets an attribute whose value is HTML-escaped when rendered.
	SetAttr(name, value string)
	// SetAttrUnsafe sets an attribute whose value is rendered verbatim.
	SetAttrUnsafe(name, value string)
}

// Apply sets each attribute on el, in order. Attributes with an empty Name are skipped.
// Generated elements expose the same behaviour as a chainable Attrs method.
func Apply(el Element, attrs ...Attribute) {
	for _, a := range attrs {
		switch {
		case a.Name == "":
			continue
		case a.Unsafe:
			el.SetAttrUnsafe(a.Name, a.Value)
		default:
			el.SetAttr(a.Name, a.Value)
		}
	}
}
