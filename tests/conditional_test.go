package tests

import (
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func TestXxxIfFalseRendersNothing(t *testing.T) {
	expect(t, html.DivIf(false, html.P().Text("x")), "")
	expect(t, html.SpanIf(false), "")
	expect(t, html.BrIf(false), "")
	expect(t, html.InputIf(false).Id("x"), "")
	expect(t, html.Div(html.PIf(false), html.BrIf(false)), "<div></div>")
}

func TestXxxIfTrueRenders(t *testing.T) {
	expect(t, html.DivIf(true, html.P().Text("x")), "<div><p>x</p></div>")
	expect(t, html.BrIf(true), "<br>")
}

func TestDivTernary(t *testing.T) {
	expect(t, html.DivTernary(true, html.B().Text("a"), html.I().Text("b")), "<div><b>a</b></div>")
	expect(t, html.DivTernary(false, html.B().Text("a"), html.I().Text("b")), "<div><i>b</i></div>")
}

func TestIf(t *testing.T) {
	expect(t, html.Div().If(true, html.Br()), "<div><br></div>")
	expect(t, html.Div().If(false, html.Br()), "<div></div>")
}

func TestIfElse(t *testing.T) {
	expect(t, html.Div().IfElse(true, html.B(), html.I()), "<div><b></b></div>")
	expect(t, html.Div().IfElse(false, html.B(), html.I()), "<div><i></i></div>")
}

func TestIfThen(t *testing.T) {
	add := func(d *html.DivElement) { d.Id("x").Text("hi") }
	expect(t, html.Div().IfThen(true, add), `<div id="x">hi</div>`)
	expect(t, html.Div().IfThen(false, add), "<div></div>")
}

func TestWith(t *testing.T) {
	d := html.Div().With(func(d *html.DivElement) {
		d.Class("a").Children(htemel.Text("t"))
	})
	expect(t, d, `<div class="a">t</div>`)
}
