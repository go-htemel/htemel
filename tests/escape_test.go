package tests

import (
	"bytes"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func render(t *testing.T, n htemel.Node) string {
	t.Helper()
	buf := &bytes.Buffer{}
	if err := n.Render(buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return buf.String()
}

func expect(t *testing.T, n htemel.Node, want string) {
	t.Helper()
	if got := render(t, n); got != want {
		t.Errorf("\nwant: %s\n got: %s", want, got)
	}
}

func TestAttributeEscaping(t *testing.T) {
	expect(t, html.Div().Title(`"><script>alert(1)</script>`),
		`<div title="&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt;"></div>`)
	expect(t, html.A().Href(`/x?a=1&b="2"`).Class(`c"d`),
		`<a href="/x?a=1&amp;b=&#34;2&#34;" class="c&#34;d"></a>`)
	expect(t, html.Div().Data("x", `<"&>`), `<div data-x="&lt;&#34;&amp;&gt;"></div>`)
	expect(t, html.Div().DataUnsafe("x", `{"a":1}`), `<div data-x="{"a":1}"></div>`)
	expect(t, html.Div().Tabindex(-1), `<div tabindex="-1"></div>`)
}

func TestGenericEscaping(t *testing.T) {
	expect(t, htemel.Generic("x-el", map[string]any{
		"title":  `"><b>`,
		"hidden": nil,
		"on":     true,
		"off":    false,
		"n":      3,
	}, htemel.Text("<hi>")),
		`<x-el hidden n="3" on title="&#34;&gt;&lt;b&gt;">&lt;hi&gt;</x-el>`)
	expect(t, htemel.GenericVoid("br", nil), `<br>`)
}

func TestAttributeOrder(t *testing.T) {
	d := html.Div().Id("a").Class("b").Title("c").Data("k", "v").Lang("en").Role("note")
	want := `<div id="a" class="b" title="c" data-k="v" lang="en" role="note"></div>`
	for i := 0; i < 50; i++ {
		expect(t, d, want)
	}
	// Re-setting a key updates the value in place.
	expect(t, html.Div().Id("a").Class("b").Id("z"), `<div id="z" class="b"></div>`)
}

func TestBooleanAttributes(t *testing.T) {
	expect(t, html.Input().Autofocus(true), `<input autofocus>`)
	expect(t, html.Input().Autofocus(false), `<input>`)
	expect(t, html.Input().Disabled(true).Name("n").Disabled(false), `<input name="n">`)
	expect(t, html.Script().Src("a.js").Defer(true).Async(false), `<script src="a.js" defer></script>`)
}

func TestClassHelpers(t *testing.T) {
	expect(t, html.Div().AddClass("a", "b"), `<div class="a b"></div>`)
	expect(t, html.Div().ToggleClass("a", true), `<div class="a"></div>`)
	expect(t, html.Div().ToggleClass("a", false), `<div></div>`)
	expect(t, html.Div().Class("a", "b").ToggleClass("a", false).AddClass("c"), `<div class="b c"></div>`)
	expect(t, html.Div().Class("a").ToggleClass("a", false), `<div></div>`)
}

func TestRawTextElements(t *testing.T) {
	expect(t, html.Script().Text("if (a < b && c > d) {}"), `<script>if (a < b && c > d) {}</script>`)
	expect(t, html.Script().Textf("var s = %q;", "</SCRIPT><b>"), `<script>var s = "<\/SCRIPT><b>";</script>`)
	expect(t, html.Style().Text("a > b { content: '&'; }"), `<style>a > b { content: '&'; }</style>`)
	expect(t, html.Style().Text("</style><b>"), `<style><\/style><b></style>`)
	// Non raw-text elements keep escaping.
	expect(t, html.P().Text("a < b"), `<p>a &lt; b</p>`)
	expect(t, html.Textarea().Text("</textarea>"), `<textarea>&lt;/textarea&gt;</textarea>`)
}

func TestTernary(t *testing.T) {
	expect(t, html.DivTernary(true, htemel.Text("yes"), htemel.Text("no")), `<div>yes</div>`)
	expect(t, html.DivTernary(false, htemel.Text("yes"), htemel.Text("no")), `<div>no</div>`)
}

func TestAttrEscapeHatch(t *testing.T) {
	expect(t, html.Button().Attr("hx-post", "/x?a=1&b=2").Attr("data-on-click", `$count++; "x"`),
		`<button hx-post="/x?a=1&amp;b=2" data-on-click="$count++; &#34;x&#34;"></button>`)
	expect(t, html.Div().AttrUnsafe("x-data", `{"open":false}`), `<div x-data="{"open":false}"></div>`)
	expect(t, html.Div().BoolAttr("hx-boost", true).BoolAttr("x-cloak", false), `<div hx-boost></div>`)
	expect(t, html.Div().BoolAttr("x", true).BoolAttr("x", false), `<div></div>`)
}

func TestDoctype(t *testing.T) {
	expect(t, htemel.Group(htemel.Doctype(), html.Html()), `<!DOCTYPE html><html></html>`)
}
