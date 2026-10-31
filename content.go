package htemel

import "io"

// Content categories.
//
// The generated element packages check the HTML content model at compile time: every element
// constructor and Children method accepts only the children its specification permits, expressed
// as one of the interfaces below (or, for elements whose model lists specific tags, an interface
// generated in the element package, such as html.UlChild).
//
// Each interface has an unexported marker method, so a type opts in by embedding the matching
// exported marker struct (FlowMarker, PhrasingMarker, ...). Generated elements embed the markers
// of their categories, so html.Span() is a PhrasingContent and a FlowContent, while html.Div() is
// only a FlowContent. A component type of your own can do the same:
//
//	type Badge struct {
//		htemel.PhrasingMarker // usable wherever phrasing content is permitted
//		label string
//	}
//
// Only the categories that appear in content models have interfaces. The checks cover the
// permitted children of each element; descendant restrictions (no interactive content inside
// <a> or <button>, no <form> inside <form>), ordering, cardinality and attribute-dependent rules
// are not expressed in the type system. Dynamic content built with Group, Each, If, Generic or
// Unchecked is trusted and accepted anywhere (see Wildcard).

// FlowContent is implemented by flow content: most elements used in the body of a document.
type FlowContent interface {
	Node
	flowContent()
}

// PhrasingContent is implemented by phrasing content: text and the elements that mark it up.
// Phrasing content is also flow content, so a PhrasingContent is a FlowContent.
type PhrasingContent interface {
	FlowContent
	phrasingContent()
}

// MetadataContent is implemented by metadata content, such as the children of <head>.
type MetadataContent interface {
	Node
	metadataContent()
}

// SectioningContent is implemented by sectioning content (article, aside, nav, section).
type SectioningContent interface {
	Node
	sectioningContent()
}

// HeadingContent is implemented by heading content (h1-h6, hgroup).
type HeadingContent interface {
	Node
	headingContent()
}

// EmbeddedContent is implemented by embedded content (img, video, iframe, svg, math, ...).
type EmbeddedContent interface {
	Node
	embeddedContent()
}

// InteractiveContent is implemented by interactive content (a, button, input, ...). Several
// elements are interactive only with certain attributes; the category is static here.
type InteractiveContent interface {
	Node
	interactiveContent()
}

// PalpableContent is implemented by palpable content.
type PalpableContent interface {
	Node
	palpableContent()
}

// ScriptSupportingContent is implemented by script-supporting elements (script, template).
type ScriptSupportingContent interface {
	Node
	scriptSupportingContent()
}

// TextContent is implemented by text nodes (Text, TextUnsafe) and is the child type of elements
// whose content model is text only, such as <title>, <script> and <textarea>.
type TextContent interface {
	Node
	textContent()
}

// FlowMarker is embedded by types that are flow content. It has no fields and no runtime cost.
type FlowMarker struct{}

func (FlowMarker) flowContent() {}

// PhrasingMarker is embedded by types that are phrasing content. It embeds FlowMarker, because
// phrasing content is also flow content.
type PhrasingMarker struct{ FlowMarker }

func (PhrasingMarker) phrasingContent() {}

// MetadataMarker is embedded by types that are metadata content.
type MetadataMarker struct{}

func (MetadataMarker) metadataContent() {}

// SectioningMarker is embedded by types that are sectioning content.
type SectioningMarker struct{}

func (SectioningMarker) sectioningContent() {}

// HeadingMarker is embedded by types that are heading content.
type HeadingMarker struct{}

func (HeadingMarker) headingContent() {}

// EmbeddedMarker is embedded by types that are embedded content.
type EmbeddedMarker struct{}

func (EmbeddedMarker) embeddedContent() {}

// InteractiveMarker is embedded by types that are interactive content.
type InteractiveMarker struct{}

func (InteractiveMarker) interactiveContent() {}

// PalpableMarker is embedded by types that are palpable content.
type PalpableMarker struct{}

func (PalpableMarker) palpableContent() {}

// ScriptSupportingMarker is embedded by script-supporting types.
type ScriptSupportingMarker struct{}

func (ScriptSupportingMarker) scriptSupportingContent() {}

// TextMarker is embedded by types that are text content.
type TextMarker struct{}

func (TextMarker) textContent() {}

// Wildcard is embedded by nodes whose content is not checked: it implements every content
// category and every element-specific child interface of the generated html, svg and mathml
// packages, so a type embedding it is accepted as a child of any element. Group, Generic and
// Doctype embed it; Each, If, Unchecked and friends return a NodeFunc, which is also a wildcard.
//
// Embed it in your own component type to opt out of checking for that type.
type Wildcard struct{}

func (Wildcard) flowContent()             {}
func (Wildcard) phrasingContent()         {}
func (Wildcard) metadataContent()         {}
func (Wildcard) sectioningContent()       {}
func (Wildcard) headingContent()          {}
func (Wildcard) embeddedContent()         {}
func (Wildcard) interactiveContent()      {}
func (Wildcard) palpableContent()         {}
func (Wildcard) scriptSupportingContent() {}
func (Wildcard) textContent()             {}

// NodeFunc is a Node implemented by a function. Like Wildcard it is accepted as a child of any
// element; it is the result type of Each, EachIndex, ForEach, ForEach2, If, IfElse and Unchecked.
type NodeFunc func(w io.Writer) error

// Render implements the Node interface by calling f.
func (f NodeFunc) Render(w io.Writer) error { return f(w) }

func (NodeFunc) flowContent()             {}
func (NodeFunc) phrasingContent()         {}
func (NodeFunc) metadataContent()         {}
func (NodeFunc) sectioningContent()       {}
func (NodeFunc) headingContent()          {}
func (NodeFunc) embeddedContent()         {}
func (NodeFunc) interactiveContent()      {}
func (NodeFunc) palpableContent()         {}
func (NodeFunc) scriptSupportingContent() {}
func (NodeFunc) textContent()             {}

// Unchecked wraps n so that it is accepted as a child of any element, bypassing the content
// model check. Use it for a Node from elsewhere (a function returning Node, a third-party
// component) whose placement you know to be valid. A nil n renders nothing.
//
//	html.Ul(htemel.Unchecked(legacyListItems()))
func Unchecked(n Node) NodeFunc {
	switch n := n.(type) {
	case nil:
		return nop
	case NodeFunc:
		return n
	}
	return n.Render
}

// nop is the NodeFunc that renders nothing.
var nop NodeFunc = func(io.Writer) error { return nil }

// The markers are only embedded by types of other packages, which the unused linter cannot see.
var (
	_ interface{ flowContent() }             = FlowMarker{}
	_ interface{ phrasingContent() }         = PhrasingMarker{}
	_ interface{ metadataContent() }         = MetadataMarker{}
	_ interface{ sectioningContent() }       = SectioningMarker{}
	_ interface{ headingContent() }          = HeadingMarker{}
	_ interface{ embeddedContent() }         = EmbeddedMarker{}
	_ interface{ interactiveContent() }      = InteractiveMarker{}
	_ interface{ palpableContent() }         = PalpableMarker{}
	_ interface{ scriptSupportingContent() } = ScriptSupportingMarker{}
	_ interface{ textContent() }             = TextMarker{}
)
