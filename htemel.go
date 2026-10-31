package htemel

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"golang.org/x/net/html"

	"github.com/go-htemel/htemel/internal/render"
)

// Node defines the interface that must be implemented in order to render elements.
type Node interface {
	Render(w io.Writer) error
}

// GroupElement is a struct that backs the Group function.
type GroupElement struct {
	children []Node
}

// Group is a generic wrapper that can be used to wrap one or more elements that may not have a suitable parent type.
func Group(children ...Node) *GroupElement {
	return &GroupElement{
		children: children,
	}
}

// Render implements the Node interface by calling Render on all child nodes.
func (e *GroupElement) Render(w io.Writer) error {
	for _, child := range e.children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	return nil
}

// GenericElement struct backs the Generic element functions.
type GenericElement struct {
	tag      string
	attrs    map[string]any
	void     bool
	children []Node
}

// Generic element is provided as an escape-hatch for when the provided generated elements are not sufficient.
// Attributes can be passed as a map of strings with "any" type.
// The underlying type should implement the fmt.Stringer interface for predictable rendering.
func Generic(tag string, attrs map[string]any, children ...Node) *GenericElement {
	return &GenericElement{
		tag:      tag,
		attrs:    attrs,
		children: children,
	}
}

// GenericVoid element is provided as an escape-hatch for when the provided generated elements are not sufficient.
// Attributes can be passed as a map of strings with "any" type.
// The underlying type should implement the fmt.Stringer interface for predictable rendering.
//
// Void elements are self-closing and therefore do not permit children.
func GenericVoid(tag string, attrs map[string]any) *GenericElement {
	return &GenericElement{
		tag:      tag,
		attrs:    attrs,
		void:     true,
		children: nil,
	}
}

// Render implements the Node interface.
//
// Attributes are rendered in sorted key order (the input is a map, so there is no insertion order to preserve).
// Values are HTML-escaped. A nil or true value renders a bare attribute; a false value omits it.
// Other values are formatted with fmt.Sprint.
func (e *GenericElement) Render(w io.Writer) error {
	var sb strings.Builder
	sb.WriteString("<" + e.tag)

	keys := make([]string, 0, len(e.attrs))
	for key := range e.attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		switch val := e.attrs[key].(type) {
		case nil:
			render.WriteAttr(&sb, key, "", false)
		case bool:
			if val {
				render.WriteAttr(&sb, key, "", false)
			}
		default:
			str := fmt.Sprint(val)
			if str == "" {
				// Keep an explicit empty value distinct from a bare attribute.
				sb.WriteString(" " + key + `=""`)
			} else {
				render.WriteAttr(&sb, key, str, false)
			}
		}
	}

	sb.WriteByte('>')

	if !e.void {
		for _, child := range e.children {
			if err := child.Render(&sb); err != nil {
				return err
			}
		}
		sb.WriteString("</" + e.tag + ">")
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// DoctypeElement renders the HTML5 document type declaration.
type DoctypeElement struct{}

// Doctype returns a node that renders <!DOCTYPE html>.
func Doctype() DoctypeElement { return DoctypeElement{} }

// Render implements the Node interface.
func (DoctypeElement) Render(w io.Writer) error {
	_, err := io.WriteString(w, "<!DOCTYPE html>")
	return err
}

// TextElement backs the Text and TextUnsafe functions.
type TextElement struct {
	text     string
	children []Node
}

// TextUnsafe renders text verbatim, without HTML escaping. Never pass untrusted input.
func TextUnsafe(text string, children ...Node) *TextElement {
	return &TextElement{
		text:     text,
		children: children,
	}
}

// Text renders HTML-escaped text.
func Text(text string, children ...Node) *TextElement {
	return &TextElement{
		text:     html.EscapeString(text),
		children: children,
	}
}

func (e *TextElement) Render(w io.Writer) error {
	if _, err := w.Write([]byte(e.text)); err != nil {
		return err
	}

	for _, child := range e.children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	return nil
}
