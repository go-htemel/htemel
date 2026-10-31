package tests

import (
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func TestVoidElements(t *testing.T) {
	tests := []struct {
		tag  string
		node htemel.Node
	}{
		{"base", html.Base()},
		{"link", html.Link()},
		{"meta", html.Meta()},
		{"hr", html.Hr()},
		{"br", html.Br()},
		{"wbr", html.Wbr()},
		{"source", html.Source()},
		{"img", html.Img()},
		{"embed", html.Embed()},
		{"track", html.Track()},
		{"area", html.Area()},
		{"col", html.Col()},
		{"input", html.Input()},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			expect(t, tt.node, "<"+tt.tag+">")
		})
	}
}

func TestVoidElementWithAttributes(t *testing.T) {
	expect(t, html.Br().Class("a"), `<br class="a">`)
	expect(t, html.Hr().Id("x"), `<hr id="x">`)
}

func TestGroup(t *testing.T) {
	expect(t, htemel.Group(), "")
	expect(t, htemel.Group(html.Br(), htemel.Text("a<b"), html.Hr()), "<br>a&lt;b<hr>")
	expect(t, htemel.Group(htemel.Group(html.Br()), html.P().Text("x")), "<br><p>x</p>")
}

func TestGeneric(t *testing.T) {
	expect(t, htemel.Generic("x-el", nil), "<x-el></x-el>")
	expect(t, htemel.Generic("x-el", map[string]any{"b": "2", "a": 1}, htemel.Text("hi")),
		`<x-el a="1" b="2">hi</x-el>`)
	expect(t, htemel.Generic("x-el", map[string]any{"on": true, "off": false, "bare": nil}),
		`<x-el bare on></x-el>`)
	expect(t, htemel.Generic("x-el", map[string]any{"v": ""}), `<x-el v=""></x-el>`)
}

func TestGenericVoid(t *testing.T) {
	expect(t, htemel.GenericVoid("x-void", nil), "<x-void>")
	expect(t, htemel.GenericVoid("x-void", map[string]any{"k": "v"}), `<x-void k="v">`)
}
