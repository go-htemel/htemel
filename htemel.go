package htemel

//go:generate go run ./cmd/genelements

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"

	"github.com/go-htemel/htemel/internal/render"
)

// Node defines the interface that must be implemented in order to render elements.
//
// Render writes the node's HTML directly to w as it walks the tree. It does not buffer:
// a page becomes many small writes, which is cheap for an in-memory writer such as
// *bytes.Buffer or *strings.Builder but expensive for one where each Write costs a syscall or
// a network flush, such as an *os.File, os.Stdout or some http.ResponseWriter implementations.
// Callers who care should use the package-level Render (which buffers and flushes), RenderString
// or RenderBytes, or pass their own buffered writer (for example a *bufio.Writer they flush).
//
// If Render returns an error, part of the output may already have been written to w.
type Node interface {
	Render(w io.Writer) error
}

// bufSize is the size of the pooled bufio.Writers used by Render.
const bufSize = 8 << 10

// maxPooledBuffer caps the capacity of buffers returned to the pool so one huge page does not
// pin a large allocation forever.
const maxPooledBuffer = 1 << 20

var (
	writerPool = sync.Pool{New: func() any { return bufio.NewWriterSize(nil, bufSize) }}
	bufferPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

// Render renders n to w through a pooled buffer and flushes it, so the many small writes made
// while walking the tree reach w as a few large writes. Use it when w is expensive to write to,
// such as an http.ResponseWriter, *os.File or os.Stdout.
//
// In-memory writers (*bytes.Buffer, *strings.Builder) are written to directly, since buffering
// them again would only add a copy.
//
// If rendering fails, output still held in the buffer is discarded rather than flushed, and the
// error is returned; output from earlier flushes may already have reached w.
func Render(w io.Writer, n Node) error {
	switch w.(type) {
	case *bytes.Buffer, *strings.Builder:
		return n.Render(w)
	}

	bw := writerPool.Get().(*bufio.Writer)
	bw.Reset(w)
	err := n.Render(bw)
	if err == nil {
		err = bw.Flush()
	}
	bw.Reset(nil) // drop the reference to w (and any unflushed output) before pooling
	writerPool.Put(bw)
	return err
}

// renderToBuffer renders n into a pooled buffer and passes its contents to fn before the buffer
// is returned to the pool.
func renderToBuffer(n Node, fn func(b []byte)) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	err := n.Render(buf)
	if err == nil {
		fn(buf.Bytes())
	}
	if buf.Cap() <= maxPooledBuffer {
		bufferPool.Put(buf)
	}
	return err
}

// RenderString renders n and returns the HTML as a string. It reuses pooled buffers, so the
// returned string is the only allocation in the common case.
func RenderString(n Node) (string, error) {
	var s string
	err := renderToBuffer(n, func(b []byte) { s = string(b) })
	return s, err
}

// RenderBytes renders n and returns the HTML as a newly allocated byte slice owned by the caller.
func RenderBytes(n Node) ([]byte, error) {
	var out []byte
	err := renderToBuffer(n, func(b []byte) { out = bytes.Clone(b) })
	return out, err
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
// Other values are formatted as fmt.Sprint would format them.
func (e *GenericElement) Render(w io.Writer) error {
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, e.tag); err != nil {
		return err
	}

	keys := make([]string, 0, len(e.attrs))
	for key := range e.attrs {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		var err error
		switch val := e.attrs[key].(type) {
		case nil:
			_, err = render.WriteAttr(w, key, "", false)
		case bool:
			if val {
				_, err = render.WriteAttr(w, key, "", false)
			}
		default:
			err = writeGenericAttr(w, key, genericValue(val))
		}
		if err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}

	if e.void {
		return nil
	}

	for _, child := range e.children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "</"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, e.tag); err != nil {
		return err
	}
	_, err := io.WriteString(w, ">")
	return err
}

// genericValue formats an attribute value exactly as fmt.Sprint would, skipping fmt for the
// common string and int cases.
func genericValue(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprint(v)
	}
}

// writeGenericAttr writes a non-boolean Generic attribute. Unlike render.WriteAttr, an empty
// value renders as key="" so it stays distinct from a bare attribute.
func writeGenericAttr(w io.Writer, key, value string) error {
	if value != "" {
		_, err := render.WriteAttr(w, key, value, false)
		return err
	}
	if _, err := io.WriteString(w, " "); err != nil {
		return err
	}
	if _, err := io.WriteString(w, key); err != nil {
		return err
	}
	_, err := io.WriteString(w, `=""`)
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

// Text renders HTML-escaped text. The text is escaped once, when the node is built.
func Text(text string, children ...Node) *TextElement {
	return &TextElement{
		text:     html.EscapeString(text),
		children: children,
	}
}

// Render implements the Node interface.
func (e *TextElement) Render(w io.Writer) error {
	if _, err := io.WriteString(w, e.text); err != nil {
		return err
	}

	for _, child := range e.children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	return nil
}
