package tests

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	nethtml "golang.org/x/net/html"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func TestAttrEscapingMatchesEscapeString(t *testing.T) {
	values := []string{
		"", "plain", "&", "'", "<", ">", `"`, "\r",
		`a&b'c<d>e"f` + "\rg", "&&<<>>", "trailing&", "&leading", "unicode é <ü>", "\r\n",
	}
	for _, v := range values {
		var buf bytes.Buffer
		if err := html.Div().Attr("x", v).Render(&buf); err != nil {
			t.Fatal(err)
		}
		want := `<div x></div>`
		if v != "" {
			want = `<div x="` + nethtml.EscapeString(v) + `"></div>`
		}
		if buf.String() != want {
			t.Errorf("value %q: got %s, want %s", v, buf.String(), want)
		}
	}
}

func TestRenderHelpersMatchDirectRender(t *testing.T) {
	for name, n := range map[string]htemel.Node{
		"realistic": realisticPage(),
		"deep":      deepTree(8, 20),
		"generic":   htemel.Generic("x-el", map[string]any{"b": 3, "a": "q&\"", "c": true}, htemel.Text("<t>")),
	} {
		var direct bytes.Buffer
		if err := n.Render(&direct); err != nil {
			t.Fatal(err)
		}

		s, err := htemel.RenderString(n)
		if err != nil || s != direct.String() {
			t.Errorf("%s: RenderString mismatch (err %v)", name, err)
		}

		b, err := htemel.RenderBytes(n)
		if err != nil || string(b) != direct.String() {
			t.Errorf("%s: RenderBytes mismatch (err %v)", name, err)
		}

		// A writer that is neither *bytes.Buffer nor *strings.Builder goes through the pooled bufio.Writer.
		var sink onlyWriter
		if err := htemel.Render(&sink, n); err != nil || sink.String() != direct.String() {
			t.Errorf("%s: Render mismatch (err %v)", name, err)
		}

		var sb strings.Builder
		if err := htemel.Render(&sb, n); err != nil || sb.String() != direct.String() {
			t.Errorf("%s: Render into strings.Builder mismatch (err %v)", name, err)
		}
	}
}

func TestRenderLargerThanBuffer(t *testing.T) {
	// Larger than the 8 KiB pooled bufio.Writer, so Render must flush more than once.
	d := html.Div()
	for i := 0; i < 2000; i++ {
		d.Children(html.Span().Class("c").Text("x&y"))
	}
	var direct bytes.Buffer
	if err := d.Render(&direct); err != nil {
		t.Fatal(err)
	}
	var sink onlyWriter
	if err := htemel.Render(&sink, d); err != nil || sink.String() != direct.String() {
		t.Errorf("large Render mismatch (err %v, len %d vs %d)", err, sink.buf.Len(), direct.Len())
	}
}

func TestGenericRender(t *testing.T) {
	got, err := htemel.RenderString(htemel.Generic("x-el",
		map[string]any{"b": 3, "a": "q&\"", "c": true, "d": false, "e": "", "f": nil, "g": 1.5},
		htemel.Text("<t>")))
	if err != nil {
		t.Fatal(err)
	}
	want := `<x-el a="q&amp;&#34;" b="3" c e="" f g="1.5">&lt;t&gt;</x-el>`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRenderPropagatesErrors(t *testing.T) {
	boom := errors.New("boom")
	page := html.Div(html.P().Text("before"), failNode{err: boom})

	if err := htemel.Render(&onlyWriter{}, page); !errors.Is(err, boom) {
		t.Errorf("Render: got %v, want boom", err)
	}
	if _, err := htemel.RenderString(page); !errors.Is(err, boom) {
		t.Errorf("RenderString: got %v, want boom", err)
	}
	if _, err := htemel.RenderBytes(page); !errors.Is(err, boom) {
		t.Errorf("RenderBytes: got %v, want boom", err)
	}

	// Write errors from the destination surface on flush.
	if err := htemel.Render(errWriter{boom}, html.Div()); !errors.Is(err, boom) {
		t.Errorf("Render with failing writer: got %v, want boom", err)
	}
	if err := html.Div().Class("c").Render(errWriter{boom}); !errors.Is(err, boom) {
		t.Errorf("direct Render with failing writer: got %v, want boom", err)
	}
}

// onlyWriter hides bytes.Buffer's concrete type so htemel.Render takes the buffered path.
type onlyWriter struct{ buf bytes.Buffer }

func (w *onlyWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *onlyWriter) String() string              { return w.buf.String() }

// failNode is a custom component; embedding FlowMarker makes it flow content.
type failNode struct {
	htemel.FlowMarker
	err error
}

func (f failNode) Render(io.Writer) error { return f.err }

type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }
