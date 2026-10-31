package datastar_test

import (
	"fmt"
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

func show(n htemel.Node) {
	s, err := htemel.RenderString(n)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}

// The example from the package documentation.
func Example() {
	show(html.Div(
		html.Input().Type(html.InputTypeText).Attrs(datastar.Bind("name")),
		html.Span().Attrs(datastar.Text("$name")),
		html.Button().Text("Save").Attrs(
			datastar.OnClick(datastar.Post("/save")),
			datastar.Indicator("saving"),
			datastar.Attr("disabled", "$saving"),
		),
	).Attrs(datastar.Signals(map[string]any{"name": "", "saving": false})))
	// Output:
	// <div data-signals="{&#34;name&#34;:&#34;&#34;,&#34;saving&#34;:false}"><input type="text" data-bind:name><span data-text="$name"></span><button data-on:click="@post(&#39;/save&#39;)" data-indicator:saving data-attr:disabled="$saving">Save</button></div>
}

func ExampleSignals() {
	type form struct {
		Count int    `json:"count"`
		Query string `json:"query"`
	}
	show(html.Div().Attrs(datastar.Signals(form{Count: 1, Query: "go"}, datastar.IfMissing())))
	// Output:
	// <div data-signals__ifmissing="{&#34;count&#34;:1,&#34;query&#34;:&#34;go&#34;}"></div>
}

func ExampleSignal() {
	show(html.Div().Attrs(
		datastar.Signal("userName", "''"),
		datastar.Computed("upper", "$userName.toUpperCase()"),
	))
	// Output:
	// <div data-signals:user-name="&#39;&#39;" data-computed:upper="$userName.toUpperCase()"></div>
}

func ExampleOn() {
	show(html.Input().Attrs(
		datastar.On("input", datastar.Get("/search"), datastar.Debounce(300*time.Millisecond).Leading()),
		datastar.On("keydown", "$open = false", datastar.Window()),
	))
	// Output:
	// <input data-on:input__debounce.300ms.leading="@get(&#39;/search&#39;)" data-on:keydown__window="$open = false">
}

func ExampleOnInterval() {
	show(html.Div().Attrs(datastar.OnInterval(datastar.Get("/poll"), datastar.Duration(5*time.Second).Leading())))
	// Output:
	// <div data-on-interval__duration.5000ms.leading="@get(&#39;/poll&#39;)"></div>
}

func ExampleGet() {
	fmt.Println(datastar.Get("/items/it's"))
	fmt.Println(datastar.Post("/save", "{contentType: 'form'}"))
	fmt.Println(datastar.Delete("/items/1"))
	// Output:
	// @get('/items/it\'s')
	// @post('/save', {contentType: 'form'})
	// @delete('/items/1')
}

func ExampleClass() {
	show(html.Li().Text("Buy milk").Attrs(
		datastar.Class("done", "$done"),
		datastar.Show("!$hidden"),
		datastar.OnClick("$done = !$done", datastar.Prevent(), datastar.Once()),
	))
	// Output:
	// <li data-class:done="$done" data-show="!$hidden" data-on:click__prevent__once="$done = !$done">Buy milk</li>
}
