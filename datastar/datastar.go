package datastar

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/go-htemel/htemel"
)

// Custom builds a Datastar attribute: data-<plugin>, then :<key> when key is non-empty, then the
// modifiers. key is written unchanged. An empty value renders a bare attribute.
//
//	Custom("on", "my-event", "@get('/x')", Once()) // data-on:my-event__once="@get('/x')"
func Custom(plugin, key, value string, mods ...Modifier) htemel.Attribute {
	var b strings.Builder
	b.WriteString("data-")
	b.WriteString(plugin)
	if key != "" {
		b.WriteByte(':')
		b.WriteString(key)
	}
	for _, m := range mods {
		b.WriteString(m.String())
	}
	return htemel.Attr(b.String(), value)
}

// signalKey writes a camelCase signal name in kebab-case, which Datastar converts back to
// camelCase: userName -> user-name. Other characters, including dots in nested paths, are kept.
func signalKey(name string) string {
	if strings.IndexFunc(name, unicode.IsUpper) < 0 {
		return name
	}
	var b strings.Builder
	b.Grow(len(name) + 4)
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('-')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Signals

// Signals returns data-signals with v encoded as JSON, merging its fields into the signals:
//
//	Signals(map[string]any{"count": 0, "user": map[string]any{"name": ""}})
//
// v is typically a map or a struct (encoded with its json tags). Signals panics if v cannot be
// encoded as JSON (a channel, a function, a cycle); such values are programming errors, like a bad
// template passed to template.Must. Use MarshalSignals to handle the error instead, or
// SignalsJSON to pass an object literal written by hand.
func Signals(v any, mods ...Modifier) htemel.Attribute {
	a, err := MarshalSignals(v, mods...)
	if err != nil {
		panic(err)
	}
	return a
}

// MarshalSignals is Signals, returning the JSON encoding error instead of panicking.
func MarshalSignals(v any, mods ...Modifier) (htemel.Attribute, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return htemel.Attribute{}, fmt.Errorf("datastar: encoding signals: %w", err)
	}
	return Custom("signals", "", string(b), mods...), nil
}

// SignalsJSON returns data-signals with obj used verbatim (it is still HTML-escaped when
// rendered). obj is a JavaScript object literal such as "{count: 0}".
func SignalsJSON(obj string, mods ...Modifier) htemel.Attribute {
	return Custom("signals", "", obj, mods...)
}

// Signal returns data-signals:<name>, setting one signal to the result of expr:
// Signal("count", "0") renders data-signals:count="0". See the package doc for how name is cased.
func Signal(name, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("signals", signalKey(name), expr, mods...)
}

// Computed returns data-computed:<name>, a read-only signal derived from expr.
func Computed(name, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("computed", signalKey(name), expr, mods...)
}

// Bind returns data-bind:<signal>, two-way binding the element's value to the signal (creating it
// if needed).
func Bind(signal string, mods ...Modifier) htemel.Attribute {
	return Custom("bind", signalKey(signal), "", mods...)
}

// Ref returns data-ref:<signal>, storing a reference to the element in the signal.
func Ref(signal string, mods ...Modifier) htemel.Attribute {
	return Custom("ref", signalKey(signal), "", mods...)
}

// Indicator returns data-indicator:<signal>, a signal that is true while a fetch started from this
// element is in flight.
func Indicator(signal string, mods ...Modifier) htemel.Attribute {
	return Custom("indicator", signalKey(signal), "", mods...)
}

// JSONSignals returns data-json-signals, which renders the signals as JSON text inside the element
// (useful for debugging). filter is an optional object such as "{include: /^user/}"; pass "" for
// all signals.
func JSONSignals(filter string, mods ...Modifier) htemel.Attribute {
	return Custom("json-signals", "", filter, mods...)
}

// DOM

// Text returns data-text, setting the element's text content to the result of expr.
func Text(expr string) htemel.Attribute { return Custom("text", "", expr) }

// Show returns data-show, showing or hiding the element based on expr.
func Show(expr string) htemel.Attribute { return Custom("show", "", expr) }

// Class returns data-class:<name>, adding the class while expr is truthy.
func Class(name, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("class", name, expr, mods...)
}

// Classes returns data-class with an object literal mapping class names to expressions:
// Classes("{hidden: $done, 'font-bold': $important}").
func Classes(obj string) htemel.Attribute { return Custom("class", "", obj) }

// Attr returns data-attr:<name>, binding the attribute to the result of expr. A false result
// removes the attribute.
func Attr(name, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("attr", name, expr, mods...)
}

// Attrs returns data-attr with an object literal mapping attribute names to expressions.
func Attrs(obj string) htemel.Attribute { return Custom("attr", "", obj) }

// Style returns data-style:<property>, binding the inline style property to the result of expr.
func Style(property, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("style", property, expr, mods...)
}

// Styles returns data-style with an object literal mapping style properties to expressions.
func Styles(obj string) htemel.Attribute { return Custom("style", "", obj) }

// Ignore returns data-ignore, which makes Datastar skip the element and its descendants (only the
// element with Self).
func Ignore(mods ...Modifier) htemel.Attribute { return Custom("ignore", "", "", mods...) }

// IgnoreMorph returns data-ignore-morph, which keeps the element and its descendants from being
// changed when the DOM is morphed by a patch.
func IgnoreMorph() htemel.Attribute { return Custom("ignore-morph", "", "") }

// PreserveAttr returns data-preserve-attr, keeping the named attributes' client-side values when
// the element is morphed.
func PreserveAttr(names ...string) htemel.Attribute {
	return Custom("preserve-attr", "", strings.Join(names, " "))
}

// Expressions and events

// Effect returns data-effect, running expr whenever the signals it reads change.
func Effect(expr string) htemel.Attribute { return Custom("effect", "", expr) }

// Init returns data-init, running expr when the element is loaded into the DOM. It replaces the
// pre-v1 data-on-load.
func Init(expr string, mods ...Modifier) htemel.Attribute { return Custom("init", "", expr, mods...) }

// On returns data-on:<event>, running expr when the event fires. event is written unchanged; use
// CaseMod(CaseCamel) for camelCase custom event names.
//
//	On("input", Get("/search"), Debounce(300*time.Millisecond))
//	// data-on:input__debounce.300ms="@get('/search')"
func On(event, expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on", event, expr, mods...)
}

// OnClick is On("click", ...).
func OnClick(expr string, mods ...Modifier) htemel.Attribute { return On("click", expr, mods...) }

// OnSubmit is On("submit", ...). Datastar prevents the default form submission itself.
func OnSubmit(expr string, mods ...Modifier) htemel.Attribute { return On("submit", expr, mods...) }

// OnInput is On("input", ...).
func OnInput(expr string, mods ...Modifier) htemel.Attribute { return On("input", expr, mods...) }

// OnChange is On("change", ...).
func OnChange(expr string, mods ...Modifier) htemel.Attribute { return On("change", expr, mods...) }

// OnKeyDown is On("keydown", ...).
func OnKeyDown(expr string, mods ...Modifier) htemel.Attribute {
	return On("keydown", expr, mods...)
}

// OnIntersect returns data-on-intersect, running expr when the element enters the viewport.
func OnIntersect(expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on-intersect", "", expr, mods...)
}

// OnInterval returns data-on-interval, running expr on an interval (set it with Duration).
func OnInterval(expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on-interval", "", expr, mods...)
}

// OnSignalPatch returns data-on-signal-patch, running expr whenever signals are patched.
func OnSignalPatch(expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on-signal-patch", "", expr, mods...)
}

// OnSignalPatchFilter returns data-on-signal-patch-filter, limiting which signals trigger
// data-on-signal-patch, for example "{include: /^counter$/}".
func OnSignalPatchFilter(filter string) htemel.Attribute {
	return Custom("on-signal-patch-filter", "", filter)
}

// Datastar Pro attributes. These need the Pro build of Datastar.

// Persist returns data-persist, persisting signals in localStorage (sessionStorage with Session).
// filter is an optional object such as "{include: /^draft/}"; pass "" for all signals. Pro.
func Persist(filter string, mods ...Modifier) htemel.Attribute {
	return Custom("persist", "", filter, mods...)
}

// QueryString returns data-query-string, syncing signals with the URL query string. Pro.
func QueryString(filter string, mods ...Modifier) htemel.Attribute {
	return Custom("query-string", "", filter, mods...)
}

// ReplaceURL returns data-replace-url, replacing the browser URL with the result of expr without
// reloading. Pro.
func ReplaceURL(expr string) htemel.Attribute { return Custom("replace-url", "", expr) }

// ScrollIntoView returns data-scroll-into-view, scrolling the element into view when it is
// loaded. Pro.
func ScrollIntoView(mods ...Modifier) htemel.Attribute {
	return Custom("scroll-into-view", "", "", mods...)
}

// CustomValidity returns data-custom-validity, setting the element's validation message to the
// result of expr (an empty string means valid). Pro.
func CustomValidity(expr string) htemel.Attribute { return Custom("custom-validity", "", expr) }

// OnRAF returns data-on-raf, running expr on every animation frame. Pro.
func OnRAF(expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on-raf", "", expr, mods...)
}

// OnResize returns data-on-resize, running expr when the element is resized. Pro.
func OnResize(expr string, mods ...Modifier) htemel.Attribute {
	return Custom("on-resize", "", expr, mods...)
}
