// Package datastar provides typed helpers for Datastar (https://data-star.dev) attributes and
// action expressions, for use with htemel's html and svg element builders.
//
// Every attribute helper returns an htemel.Attribute, which any generated element accepts through
// its Attrs method:
//
//	html.Div(
//		html.Input().Type(html.InputTypeText).Attrs(datastar.Bind("name")),
//		html.Span().Attrs(datastar.Text("$name")),
//		html.Button().Text("Save").Attrs(
//			datastar.OnClick(datastar.Post("/save")),
//			datastar.Indicator("saving"),
//			datastar.Attr("disabled", "$saving"),
//		),
//	).Attrs(datastar.Signals(map[string]any{"name": "", "saving": false}))
//
// renders (on one line, wrapped here for reading):
//
//	<div data-signals="{&#34;name&#34;:&#34;&#34;,&#34;saving&#34;:false}">
//	<input type="text" data-bind:name><span data-text="$name"></span>
//	<button data-on:click="@post(&#39;/save&#39;)" data-indicator:saving data-attr:disabled="$saving">Save</button>
//	</div>
//
// # Version and syntax
//
// The helpers target Datastar v1.0 and its key syntax, where an attribute's key follows a colon:
// data-on:click, data-signals:count, data-class:hidden, data-attr:disabled, data-bind:name.
// (Release candidates before v1.0.0-RC.6 used a hyphen instead, as in data-on-click; those
// versions are not supported.) The attribute list follows the v1.0 reference at
// https://data-star.dev/reference/attributes as of late 2025; check it when upgrading Datastar,
// since attribute names and modifiers have changed between releases. Helpers for
// Datastar Pro attributes are marked as such. For anything not covered here, build the attribute
// directly with htemel.Attr, or with Custom and the modifier API.
//
// # Escaping
//
// Attribute values are ordinary attribute values: they are HTML-escaped when rendered, so the JSON
// produced by Signals renders with &#34; in place of double quotes and expressions with &#39; in
// place of single quotes. That is correct HTML; the browser decodes the entities before Datastar
// reads the attribute. Datastar expressions are JavaScript, so never interpolate untrusted input
// into an expression string. Pass untrusted data as signal values through Signals (which
// JSON-encodes it) or quote it with Quote.
//
// # Signal names
//
// HTML attribute names are case-insensitive, so Datastar converts keys written in kebab-case to
// camelCase signal names. Helpers whose key is a signal name (Signal, Computed, Bind, Ref,
// Indicator) accept the camelCase name and write it in kebab-case: Signal("userName", "”")
// renders data-signals:user-name, which Datastar reads as the signal $userName. Keys that name a
// class, attribute, style property or event (Class, Attr, Style, On) are written unchanged.
//
// # Modifiers
//
// Datastar modifiers follow a double underscore, with their arguments after dots:
// data-on:input__debounce.300ms.leading__prevent. Pass them as trailing arguments:
//
//	datastar.On("input", datastar.Get("/search"), datastar.Debounce(300*time.Millisecond).Leading(), datastar.Prevent())
//
// Durations are written in whole milliseconds.
package datastar
