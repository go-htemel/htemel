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
	expect(t, html.Div().Contenteditable(), `<div contenteditable></div>`)
	expect(t, html.Div().Hidden(html.HiddenUntilFound, html.HiddenEmpty), `<div hidden></div>`)
}

func TestCharAttribute(t *testing.T) {
	expect(t, html.Div().Accesskey('k'), `<div accesskey="k"></div>`)
}
