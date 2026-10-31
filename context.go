package htemel

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
)

// ctxWriter is the writer RenderContext and Write pass down the tree: it forwards writes to the
// underlying (buffered) writer and carries the context for ContextOf.
type ctxWriter struct {
	w   io.Writer
	ctx context.Context
}

func (c *ctxWriter) Write(p []byte) (int, error) { return c.w.Write(p) }

// WriteString forwards to the underlying writer's WriteString, so io.WriteString (used by every
// generated element) does not convert to []byte.
func (c *ctxWriter) WriteString(s string) (int, error) { return io.WriteString(c.w, s) }

// ContextOf returns the context of the render that passed w to a Node's Render method: the ctx
// given to RenderContext, or the request context in Write and Handler. For any other writer
// (including a plain Render or a direct Node.Render call) it returns context.Background().
//
// It lets a component that is several levels deep reach request-scoped values (a CSP nonce, a
// locale, a trace span) or check for cancellation without threading a parameter through every
// layer:
//
//	func Nonced(src string) htemel.Node {
//		return htemel.NodeFunc(func(w io.Writer) error {
//			nonce, _ := htemel.ContextOf(w).Value(nonceKey{}).(string)
//			return html.Script().Src(src).Nonce(nonce).Render(w)
//		})
//	}
//
// Passing data as explicit parameters remains the recommended style; ContextOf is for the cases
// where that is impractical. The context is only visible through the writer htemel passes down:
// a component that renders a child into a writer of its own (a bytes.Buffer, RenderString) hides it.
func ContextOf(w io.Writer) context.Context {
	if c, ok := w.(*ctxWriter); ok {
		return c.ctx
	}
	return context.Background()
}

// RenderContext is Render with a context: it renders n to w through a pooled buffer, and every
// component in the tree can retrieve ctx with ContextOf(w).
//
// If ctx is already done, nothing is written and ctx.Err() is returned. If ctx is done or
// rendering fails, output still held in the buffer is discarded rather than flushed and the error
// is returned; as with Render, output from earlier flushes (pages over 8 KiB) may already have
// reached w. To serve HTTP with no partial output at all, use Write or Handler.
func RenderContext(ctx context.Context, w io.Writer, n Node) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	switch w.(type) {
	case *bytes.Buffer, *strings.Builder:
		return renderNode(&ctxWriter{w: w, ctx: ctx}, n)
	}

	bw := writerPool.Get().(*bufio.Writer)
	bw.Reset(w)
	err := renderNode(&ctxWriter{w: bw, ctx: ctx}, n)
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		err = bw.Flush()
	}
	bw.Reset(nil) // drop the reference to w (and any unflushed output) before pooling
	writerPool.Put(bw)
	return err
}
