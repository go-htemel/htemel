package tests

import (
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// The per-element conditionals (DivIf, DivTernary, .If, .IfElse, .IfThen) were removed in v0.4.0
// (ADR 0008). These are their replacements.

func TestPackageIfReplacesXxxIf(t *testing.T) {
	expect(t, htemel.If(false, html.Div(html.P().Text("x"))), "")
	expect(t, htemel.If(true, html.Div(html.P().Text("x"))), "<div><p>x</p></div>")
	expect(t, html.Div(htemel.If(false, html.P()), htemel.If(false, html.Br())), "<div></div>")
	expect(t, html.Div().Children(htemel.If(true, html.Br())), "<div><br></div>")
}

func TestPackageIfElseReplacesTernary(t *testing.T) {
	expect(t, html.Div(htemel.IfElse(true, html.B().Text("a"), html.I().Text("b"))), "<div><b>a</b></div>")
	expect(t, html.Div(htemel.IfElse(false, html.B().Text("a"), html.I().Text("b"))), "<div><i>b</i></div>")
}

func TestWithReplacesIfThen(t *testing.T) {
	add := func(cond bool) func(*html.DivElement) {
		return func(d *html.DivElement) {
			if cond {
				d.Id("x").Text("hi")
			}
		}
	}
	expect(t, html.Div().With(add(true)), `<div id="x">hi</div>`)
	expect(t, html.Div().With(add(false)), "<div></div>")
}

func TestWith(t *testing.T) {
	d := html.Div().With(func(d *html.DivElement) {
		d.Class("a").Children(htemel.Text("t"))
	})
	expect(t, d, `<div class="a">t</div>`)
}
