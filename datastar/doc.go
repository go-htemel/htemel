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
// The helpers target Datastar v1 and its key syntax, where an attribute's key follows a colon:
// data-on:click, data-signals:count, data-class:hidden, data-attr:disabled, data-bind:name.
// (Release candidates before v1.0.0-RC.6 used a hyphen instead, as in data-on-click; those
// versions are not supported.) Attributes that are a single plugin without a key keep their
// hyphens, as in data-on-intersect, data-on-interval, data-on-signal-patch and data-json-signals.
//
// The attribute, modifier and action helpers were verified against the Datastar v1.0.4 reference
// (https://data-star.dev/reference/attributes and /reference/actions) on 2026-10-31. Attribute
// names, modifiers and action signatures have changed between releases, so re-check the
// reference when you upgrade Datastar; the offline list in testdata/v1.0.4-attributes.txt is what
// the tests compare the helpers against. Helpers for Datastar Pro attributes and actions are
// marked "Pro" and need the Pro build. For anything not covered here, build the attribute with
// Custom and the modifier API, or directly with htemel.Attr.
//
// Datastar's attribute aliasing (data-star-* and custom prefixes) is a build-time choice of the
// bundle, not an attribute, so there is no helper; the helpers always emit data-*.
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
// Durations are written in whole milliseconds. Arguments after dots are a list, so several events
// are Event("input", "change"), which renders __event.input.change; Mod builds any other modifier.
//
// # Backend actions
//
// Get, Post, Put, Patch, Delete and Query take the URL and, optionally, a JavaScript options
// object, which RequestOptions renders:
//
//	datastar.Post("/save", datastar.RequestOptions{ContentType: datastar.ContentForm}.String())
package datastar
