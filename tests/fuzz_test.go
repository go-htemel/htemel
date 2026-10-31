package tests

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/go-htemel/htemel"
	htmlpkg "github.com/go-htemel/htemel/html"
)

var fuzzSeeds = []string{
	"",
	"plain",
	`"><script>alert(1)</script>`,
	`' onmouseover='alert(1)`,
	`" onmouseover="alert(1)`,
	`</div><div>`,
	`<img src=x onerror=alert(1)>`,
	`javascript:alert(1)`,
	`&amp; &lt; &#34; &quot;`,
	"a\tb\nc\r\nd",
	"héllo wörld 日本語 🎉",
	"\x00nul",
	"  ",
	"`backtick`",
}

func parseFragment(t *testing.T, s string) []*html.Node {
	t.Helper()
	ctx := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(s), ctx)
	if err != nil {
		t.Fatalf("parse failed for %q: %v", s, err)
	}
	return nodes
}

// roundTripSafe reports whether s survives HTML parsing unchanged. The parser normalises CR/CRLF
// to LF and drops or replaces NUL, so those inputs are excluded from value comparisons.
func roundTripSafe(s string) bool {
	return !strings.ContainsAny(s, "\r\x00")
}

func FuzzTextEscaping(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := render(t, htmlpkg.Div(htemel.Text(s)))

		nodes := parseFragment(t, out)
		if len(nodes) != 1 || nodes[0].Type != html.ElementNode || nodes[0].Data != "div" {
			t.Fatalf("input %q: expected exactly one <div>, got %d nodes; html=%s", s, len(nodes), out)
		}
		div := nodes[0]
		for c := div.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				t.Fatalf("input %q: unexpected child element <%s>; html=%s", s, c.Data, out)
			}
		}
		if !roundTripSafe(s) {
			return
		}
		var got strings.Builder
		for c := div.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				got.WriteString(c.Data)
			}
		}
		if got.String() != s {
			t.Fatalf("text round-trip mismatch: want %q got %q; html=%s", s, got.String(), out)
		}
	})
}

func FuzzAttrEscaping(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}

	cases := []struct {
		name string
		attr string
		mk   func(s string) htemel.Node
	}{
		{"Title", "title", func(s string) htemel.Node { return htmlpkg.Div().Title(s) }},
		{"Attr", "data-x", func(s string) htemel.Node { return htmlpkg.Div().Attr("data-x", s) }},
		{"Id", "id", func(s string) htemel.Node { return htmlpkg.Div().Id(s) }},
	}

	f.Fuzz(func(t *testing.T, s string) {
		for _, c := range cases {
			out := render(t, c.mk(s))

			nodes := parseFragment(t, out)
			if len(nodes) != 1 || nodes[0].Type != html.ElementNode || nodes[0].Data != "div" {
				t.Fatalf("%s(%q): expected exactly one <div>, got %d nodes; html=%s", c.name, s, len(nodes), out)
			}
			div := nodes[0]
			if div.FirstChild != nil {
				t.Fatalf("%s(%q): div should have no children; html=%s", c.name, s, out)
			}
			if len(div.Attr) != 1 || div.Attr[0].Key != c.attr {
				t.Fatalf("%s(%q): expected single attribute %q, got %+v; html=%s", c.name, s, c.attr, div.Attr, out)
			}
			if roundTripSafe(s) && div.Attr[0].Val != s {
				t.Fatalf("%s: attribute round-trip mismatch: want %q got %q; html=%s", c.name, s, div.Attr[0].Val, out)
			}
		}
	})
}
