package htemel

import (
	"bytes"
	"context"
	"log"
	"net/http"
)

// Write renders n and sends it as the body of an HTTP response. It is the safe way to serve a
// page: the whole page is rendered into a pooled buffer first, with the request's context
// available to components through ContextOf, and only written to w once rendering has succeeded.
//
// On success it sets "Content-Type: text/html; charset=utf-8" unless a Content-Type is already
// set, then writes the body in one call (the status is whatever the caller set with WriteHeader,
// or 200). On a rendering error, or if the request context is done, it writes nothing at all (no
// header, status or body) and returns the error, so the caller decides how to respond:
//
//	if err := htemel.Write(w, r, views.Page(data)); err != nil {
//		http.Error(w, "something went wrong", http.StatusInternalServerError)
//	}
//
// An error from writing the body itself (a client that went away) is returned too; by then the
// response has started. r may be nil, in which case the context is context.Background().
//
// Rendering straight to the ResponseWriter (n.Render(w), or Render for pages larger than its
// buffer) can instead leave half a page in the response, already sent with a 200 status, when a
// component fails.
func Write(w http.ResponseWriter, r *http.Request, n Node) error {
	_, err := write(w, r, n)
	return err
}

// write is Write that also reports whether the error came from rendering (nothing was written).
func write(w http.ResponseWriter, r *http.Request, n Node) (renderFailed bool, err error) {
	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledBuffer {
			bufferPool.Put(buf)
		}
	}()

	if err := renderNode(&ctxWriter{w: buf, ctx: ctx}, n); err != nil {
		return true, err
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	_, err = w.Write(buf.Bytes())
	return false, err
}

// Handler returns an http.Handler that serves the Node page(r) returns, using Write. On success
// the response is 200 with the rendered HTML. If rendering fails, no part of the page is sent: the
// error is logged with the standard log package (as net/http logs its own errors) and the client
// gets a plain-text 500 Internal Server Error. A nil Node renders an empty body.
//
//	mux.Handle("GET /about", htemel.Handler(func(r *http.Request) htemel.Node {
//		return views.About()
//	}))
//
// Use Write directly when you need a different status, headers chosen after rendering, or your
// own error page.
func Handler(page func(r *http.Request) Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderFailed, err := write(w, r, page(r))
		if err == nil || !renderFailed {
			return // a body write error means the client is gone; there is no one to tell
		}
		log.Printf("htemel: rendering %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})
}
