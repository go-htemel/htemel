package tests

import (
	"testing"

	"github.com/go-htemel/htemel/html"
)

// AllowEmpty enums take a variadic argument: none renders the bare attribute, the last one wins.
func TestAllowEmptyEnumVariadic(t *testing.T) {
	expect(t, html.Div().Hidden(), `<div hidden></div>`)
	expect(t, html.Div().Hidden(html.HiddenEmpty), `<div hidden></div>`)
	expect(t, html.Div().Hidden(html.HiddenUntilFound), `<div hidden="until-found"></div>`)
	expect(t, html.Div().Hidden(html.HiddenUntilFound, html.HiddenEmpty), `<div hidden></div>`)
}

// Rare global attributes live in html.Globals (ADR 0010).
func TestGlobals(t *testing.T) {
	expect(t, html.Div().Globals(html.Globals{Accesskey: "k"}), `<div accesskey="k"></div>`)
	expect(t, html.Div().Globals(html.Globals{Contenteditable: html.ContenteditablePlaintextOnly}), `<div contenteditable="plaintext-only"></div>`)
	// Zero fields are omitted; bools render bare; SSTs join with spaces; ints format.
	expect(t, html.Div().Globals(html.Globals{}), `<div></div>`)
	expect(t, html.Section().Globals(html.Globals{
		Itemscope: true,
		Itemtype:  []string{"https://schema.org/Person"},
		Inert:     true,
	}), `<section inert itemscope itemtype="https://schema.org/Person"></section>`)
	expect(t, html.Div().Globals(html.Globals{Headingoffset: 2, Headingreset: false}), `<div headingoffset="2"></div>`)
	// Fields apply in declaration order regardless of literal order, and chain with setters.
	expect(t, html.Input().Id("q").Globals(html.Globals{
		Spellcheck: html.SpellcheckFalse, Enterkeyhint: html.EnterkeyhintSearch, Autocapitalize: html.AutocapitalizeOff,
	}).Name("q"), `<input id="q" autocapitalize="off" enterkeyhint="search" spellcheck="false" name="q">`)
	// Values are escaped like any attribute; a later Globals call updates in place.
	expect(t, html.Span().Globals(html.Globals{Slot: `"x"`}).Globals(html.Globals{Itemprop: []string{"a", "b"}, Slot: "y"}),
		`<span slot="y" itemprop="a b"></span>`)
}
