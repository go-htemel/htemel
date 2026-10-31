package htemel_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

type ctxKey struct{}

// fromCtx renders the string stored under ctxKey in the render's context.
func fromCtx() htemel.NodeFunc {
	return func(w io.Writer) error {
		v, _ := htemel.ContextOf(w).Value(ctxKey{}).(string)
		return html.Span().Text(v).Render(w)
	}
}

// failing renders some output, then fails.
var errFail = errors.New("component failed")

func failing() htemel.NodeFunc {
	return func(w io.Writer) error {
		if _, err := io.WriteString(w, "<p>partial"); err != nil {
			return err
		}
		return errFail
	}
}

func TestRenderContextReachesComponents(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{}, "fr")
	page := html.Div(html.P(fromCtx()))
	for name, w := range map[string]io.Writer{"buffer": &bytes.Buffer{}, "builder": &strings.Builder{}, "writer": &onlyWriter{}} {
		if err := htemel.RenderContext(ctx, w, page); err != nil {
			t.Fatal(err)
		}
		if got := w.(interface{ String() string }).String(); got != "<div><p><span>fr</span></p></div>" {
			t.Errorf("%s: got %q", name, got)
		}
	}
	// Plain Render and direct Render see context.Background().
	if got, _ := htemel.RenderString(page); got != "<div><p><span></span></p></div>" {
		t.Errorf("RenderString: %q", got)
	}
	if htemel.ContextOf(&bytes.Buffer{}) != context.Background() {
		t.Error("ContextOf of a plain writer must be context.Background()")
	}
	// A nested htemel.Render inside a RenderContext tree keeps the context.
	nested := htemel.NodeFunc(func(w io.Writer) error { return htemel.Render(w, fromCtx()) })
	var b onlyWriter
	if err := htemel.RenderContext(ctx, &b, nested); err != nil || b.String() != "<span>fr</span>" {
		t.Errorf("nested Render: %q, %v", b.String(), err)
	}
}

func TestRenderContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var b onlyWriter
	if err := htemel.RenderContext(ctx, &b, html.Div()); !errors.Is(err, context.Canceled) || b.Len() != 0 {
		t.Errorf("got %v, %q", err, b.String())
	}
	// Cancelled mid-render: the buffered output is discarded.
	ctx, cancel = context.WithCancel(context.Background())
	cancelling := htemel.NodeFunc(func(w io.Writer) error { cancel(); return html.P().Render(w) })
	if err := htemel.RenderContext(ctx, &b, html.Div(cancelling)); !errors.Is(err, context.Canceled) || b.Len() != 0 {
		t.Errorf("got %v, %q", err, b.String())
	}
}

func TestRenderContextErrorDiscardsBuffer(t *testing.T) {
	var b onlyWriter
	if err := htemel.RenderContext(context.Background(), &b, html.Div(failing())); !errors.Is(err, errFail) || b.Len() != 0 {
		t.Errorf("got %v, %q", err, b.String())
	}
}

func TestWrite(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, "req"))
	rec := httptest.NewRecorder()
	if err := htemel.Write(rec, r, html.Div(fromCtx())); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "<div><span>req</span></div>" {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}

	// An existing Content-Type and status are kept.
	rec = httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/xhtml+xml")
	if err := htemel.Write(rec, nil, html.P()); err != nil {
		t.Fatal(err)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/xhtml+xml" {
		t.Errorf("Content-Type overwritten: %q", ct)
	}
}

// A failing child must produce no body bytes, no status and no headers.
func TestWriteErrorWritesNothing(t *testing.T) {
	rec := &spyWriter{ResponseRecorder: httptest.NewRecorder()}
	big := strings.Repeat("x", 64<<10) // larger than Render's buffer: no early flush either
	err := htemel.Write(rec, httptest.NewRequest(http.MethodGet, "/", nil), html.Div(html.P().Text(big), failing()))
	if !errors.Is(err, errFail) {
		t.Fatalf("err = %v", err)
	}
	if rec.Body.Len() != 0 || rec.calls != 0 || len(rec.Header()) != 0 {
		t.Errorf("wrote on error: calls=%d header=%v body=%d bytes", rec.calls, rec.Header(), rec.Body.Len())
	}
}

func TestHandler(t *testing.T) {
	h := htemel.Handler(func(r *http.Request) htemel.Node {
		if r.URL.Path == "/fail" {
			return html.Div(html.P().Text("before"), failing())
		}
		return html.H1().Text(r.URL.Path)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "<h1>/ok</h1>" || rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("ok: %d %q %v", rec.Code, rec.Body.String(), rec.Header())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fail", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("fail: status %d", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "before") || strings.Contains(body, "partial") || !strings.Contains(body, "Internal Server Error") {
		t.Errorf("fail: body %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("fail: Content-Type %q", ct)
	}
}

// onlyWriter is an io.Writer that is not one of the in-memory types Render writes to directly.
type onlyWriter struct{ bytes.Buffer }

func (w *onlyWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p) }

// spyWriter counts Write and WriteHeader calls.
type spyWriter struct {
	*httptest.ResponseRecorder
	calls int
}

func (s *spyWriter) Write(p []byte) (int, error) { s.calls++; return s.ResponseRecorder.Write(p) }
func (s *spyWriter) WriteHeader(code int)        { s.calls++; s.ResponseRecorder.WriteHeader(code) }
