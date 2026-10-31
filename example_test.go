package htemel_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func ExampleGroup() {
	// Group renders its children in order without adding a wrapping element.
	htemel.Group(
		html.P().Text("First"),
		html.P().Text("Second"),
	).Render(os.Stdout)
	// Output: <p>First</p><p>Second</p>
}

func ExampleText() {
	// Text escapes its input, so markup in user data is rendered as plain text.
	htemel.Text("<b>Tom & Jerry</b>").Render(os.Stdout)
	// Output: &lt;b&gt;Tom &amp; Jerry&lt;/b&gt;
}

func ExampleDoctype() {
	htemel.Doctype().Render(os.Stdout)
	// Output: <!DOCTYPE html>
}

func ExampleGeneric() {
	// Generic renders an element that has no generated constructor.
	// Attributes are written in sorted key order, and a true bool is a bare attribute.
	htemel.Generic("my-card", map[string]any{
		"variant": "compact",
		"open":    true,
		"hidden":  false,
	}, htemel.Text("Hello")).Render(os.Stdout)
	// Output: <my-card open variant="compact">Hello</my-card>
}

func Example_basicPage() {
	page := htemel.Group(
		htemel.Doctype(),
		html.Html().Lang("en").Children(
			html.Head().Children(
				html.Meta().Charset("UTF-8"),
				html.Title().Text("Greetings"),
			),
			html.Body().Children(
				html.H1().Class("title").Text("Hello"),
				html.P().Text("Welcome, <guest>!"),
				html.Ul().Children(
					html.Li().Text("One"),
					html.Li().Text("Two"),
				),
			),
		),
	)

	page.Render(os.Stdout)
	// Output: <!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Greetings</title></head><body><h1 class="title">Hello</h1><p>Welcome, &lt;guest&gt;!</p><ul><li>One</li><li>Two</li></ul></body></html>
}

func ExampleEach() {
	names := []string{"Ada", "Grace", "<Linus>"}
	html.Ul().Children(
		htemel.Each(names, func(n string) htemel.Node {
			return html.Li().Text(n)
		}),
	).Render(os.Stdout)
	// Output: <ul><li>Ada</li><li>Grace</li><li>&lt;Linus&gt;</li></ul>
}

func ExampleURL() {
	for _, in := range []string{"https://example.com/a", "/docs", "javascript:alert(1)"} {
		fmt.Printf("%q\n", htemel.URLOr(in, "#"))
	}
	_, err := htemel.URL("java\tscript:alert(1)")
	fmt.Println(errors.Is(err, htemel.ErrUnsafeURL))
	// Output:
	// "https://example.com/a"
	// "/docs"
	// "#"
	// true
}

func ExampleAttribute() {
	// Attributes are values, so helpers can build them without knowing the element type.
	// Every generated element applies them with Attrs; a zero Attribute is skipped.
	busy := true
	_ = html.Button().Text("Save").Attrs(
		htemel.Attr("data-on:click", "@post('/save')"),
		htemel.Attr("aria-busy", "true").If(busy),
		htemel.Attr("data-hidden", "").If(!busy),
	).Render(os.Stdout)
	// Output: <button data-on:click="@post(&#39;/save&#39;)" aria-busy="true">Save</button>
}

func ExampleApply() {
	// Apply sets attributes through the htemel.Element interface that all generated elements implement.
	var el htemel.Element = html.Span()
	htemel.Apply(el, htemel.Attr("data-text", "$name"))
	_ = el.Render(os.Stdout)
	// Output: <span data-text="$name"></span>
}
