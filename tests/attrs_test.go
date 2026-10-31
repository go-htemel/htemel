package tests

import (
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

// Every generated element implements htemel.Element, so helper packages can target it.
var (
	_ htemel.Element = (*html.DivElement)(nil)
	_ htemel.Element = (*html.BrElement)(nil) // void
	_ htemel.Element = (*html.InputElement)(nil)
	_ htemel.Element = (*svg.SvgElement)(nil)
	_ htemel.Element = (*svg.CircleElement)(nil)
)

func TestAttrsHTML(t *testing.T) {
	expect(t, html.Div().Attrs(), "<div></div>")
	expect(t, html.Div().Id("a").Attrs(
		htemel.Attr("data-x", `"<&>'`),
		htemel.AttrUnsafe("data-raw", "<b>"),
		htemel.Attr("data-bare", ""),
	), `<div id="a" data-x="&#34;&lt;&amp;&gt;&#39;" data-raw="<b>" data-bare></div>`)
	// void element
	expect(t, html.Br().Attrs(htemel.Attr("data-x", "1")), `<br data-x="1">`)
	// Attrs chains with the generated setters and keeps insertion order.
	expect(t, html.Input().Attrs(htemel.Attr("data-a", "1")).Name("q").Attrs(htemel.Attr("data-b", "2")),
		`<input data-a="1" name="q" data-b="2">`)
}

func TestAttrsOverwriteKeepsPosition(t *testing.T) {
	expect(t, html.Div().Class("x").Id("a").Attrs(htemel.Attr("class", "y")), `<div class="y" id="a"></div>`)
	// an escaped value replaced by an unsafe one and vice versa
	expect(t, html.Div().Attrs(htemel.Attr("data-v", "<"), htemel.AttrUnsafe("data-v", "<")), `<div data-v="<"></div>`)
	expect(t, html.Div().Attrs(htemel.AttrUnsafe("data-v", "<"), htemel.Attr("data-v", "<")), `<div data-v="&lt;"></div>`)
}

func TestAttrsSkipsZeroAndIf(t *testing.T) {
	expect(t, html.Div().Attrs(
		htemel.Attribute{},
		htemel.Attr("data-on", "1").If(true),
		htemel.Attr("data-off", "1").If(false),
	), `<div data-on="1"></div>`)
}

func TestAttrsSVG(t *testing.T) {
	expect(t, svg.Svg(
		svg.Circle().R("5").Attrs(htemel.Attr("data-id", "c<1>")),
	).Attrs(htemel.Attr("viewBox", "0 0 10 10")),
		`<svg viewBox="0 0 10 10"><circle r="5" data-id="c&lt;1&gt;"></circle></svg>`)
}

func TestApplyThroughInterface(t *testing.T) {
	var els []htemel.Element
	els = append(els, html.Span(), svg.Rect())
	for _, el := range els {
		htemel.Apply(el, htemel.Attr("data-k", "v"))
		el.SetAttrUnsafe("data-u", "&")
	}
	expect(t, els[0], `<span data-k="v" data-u="&"></span>`)
	expect(t, els[1], `<rect data-k="v" data-u="&"></rect>`)
}

func TestAttrsDatastar(t *testing.T) {
	expect(t, html.Button().Text("Go").Attrs(
		datastar.OnClick(datastar.Post("/go")),
		datastar.Indicator("busy"),
		datastar.Attr("disabled", "$busy"),
	), `<button data-on:click="@post(&#39;/go&#39;)" data-indicator:busy data-attr:disabled="$busy">Go</button>`)
	expect(t, svg.Circle().Attrs(datastar.Attr("r", "$radius")), `<circle data-attr:r="$radius"></circle>`)
}
