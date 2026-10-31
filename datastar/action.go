package datastar

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// Quote returns s as a single-quoted JavaScript string literal, escaping backslashes, quotes,
// line terminators and other control characters, so any Go string can be embedded safely in a
// Datastar expression: "@get(" + Quote(url) + ")".
func Quote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for _, r := range s {
		switch r {
		case '\'':
			b.WriteString(`\'`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case 0x2028, 0x2029: // line and paragraph separators end a line in older JavaScript
			b.WriteString(`\u`)
			b.WriteString(strconv.FormatInt(int64(r), 16))
		default:
			if r < 0x20 || r == 0x7f {
				b.WriteString(`\x`)
				if r < 0x10 {
					b.WriteByte('0')
				}
				b.WriteString(strconv.FormatInt(int64(r), 16))
				continue
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// action builds @name(arg0, args...) with arg0 quoted and the rest written verbatim.
func action(name, url string, args []string) string {
	var b strings.Builder
	b.WriteByte('@')
	b.WriteString(name)
	b.WriteByte('(')
	b.WriteString(Quote(url))
	for _, a := range args {
		if a == "" {
			continue
		}
		b.WriteString(", ")
		b.WriteString(a)
	}
	b.WriteByte(')')
	return b.String()
}

// Get returns the expression @get('url'), which sends a GET request to url with the signals and
// applies the response. url is quoted with Quote. Each opt is appended verbatim as a further
// argument; Datastar takes one options object, such as "{openWhenHidden: true}" or
// RequestOptions.String(). Empty opts are skipped.
//
//	Get("/search?q=a'b")                     // @get('/search?q=a\'b')
//	Post("/save", "{contentType: 'form'}")   // @post('/save', {contentType: 'form'})
func Get(url string, opts ...string) string { return action("get", url, opts) }

// Post returns the expression @post('url', opts...). See Get.
func Post(url string, opts ...string) string { return action("post", url, opts) }

// Put returns the expression @put('url', opts...). See Get.
func Put(url string, opts ...string) string { return action("put", url, opts) }

// Patch returns the expression @patch('url', opts...). See Get.
func Patch(url string, opts ...string) string { return action("patch", url, opts) }

// Delete returns the expression @delete('url', opts...). See Get.
func Delete(url string, opts ...string) string { return action("delete", url, opts) }

// Query returns the expression @query('url', opts...), a QUERY request: signals travel in the body
// like POST, but the method is safe and idempotent. See Get.
func Query(url string, opts ...string) string { return action("query", url, opts) }

// SignalFilter is a signal path filter, as taken by the filterSignals option, SetAll and ToggleAll.
// Include and Exclude are regular expression bodies without slashes; an empty field is omitted.
type SignalFilter struct {
	Include string
	Exclude string
}

// String renders the filter as a JavaScript object literal, "{include: /^user\./, exclude: /_temp$/}",
// or "" when both fields are empty. The expressions are written verbatim between slashes, so a
// slash inside one must be escaped as \/.
func (f SignalFilter) String() string {
	var parts []string
	if f.Include != "" {
		parts = append(parts, "include: /"+f.Include+"/")
	}
	if f.Exclude != "" {
		parts = append(parts, "exclude: /"+f.Exclude+"/")
	}
	if len(parts) == 0 {
		return ""
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// Content types for RequestOptions.ContentType.
const (
	ContentJSON = "json" // send the signals as JSON (the default)
	ContentForm = "form" // validate and send the closest form (or Selector); no signals are sent
)

// Retry modes for RequestOptions.Retry.
const (
	RetryAuto   = "auto"   // retry on network errors only (the default)
	RetryError  = "error"  // also retry 4xx and 5xx responses
	RetryAlways = "always" // retry every non-204 response except redirects
	RetryNever  = "never"  // never retry
)

// Request cancellation modes for RequestOptions.RequestCancellation.
const (
	CancelAuto     = "auto"     // cancel an in-flight request with the same method and URL (the default)
	CancelCleanup  = "cleanup"  // as auto, and also cancel when the element or attribute is cleaned up
	CancelDisabled = "disabled" // allow concurrent requests
)

// RequestOptions are the options object of the backend actions (Get, Post, ...). Zero fields are
// omitted so Datastar applies its defaults. Render it with String and pass it as an option:
//
//	Post("/save", RequestOptions{ContentType: ContentForm, Retry: RetryNever}.String())
//	// @post('/save', {contentType: 'form', retry: 'never'})
type RequestOptions struct {
	// ContentType is ContentJSON or ContentForm.
	ContentType string
	// FilterSignals chooses which signals are sent. By default all signals without an underscore
	// prefix are.
	FilterSignals *SignalFilter
	// Selector is a CSS selector of the form to send with ContentForm; default is the closest form.
	Selector string
	// Headers are extra request headers, written in sorted order.
	Headers map[string]string
	// OpenWhenHidden keeps the connection open while the page is hidden. Datastar defaults to
	// false for GET and true for other methods.
	OpenWhenHidden *bool
	// Payload is a JavaScript expression replacing the fetch payload.
	Payload string
	// Retry is one of the Retry constants.
	Retry string
	// RetryInterval is the first retry wait (default one second); the unit is whole milliseconds.
	RetryInterval time.Duration
	// RetryScaler multiplies the wait after each retry (default 2).
	RetryScaler float64
	// RetryMaxWait caps the wait between retries (default 30 seconds).
	RetryMaxWait time.Duration
	// RetryMaxCount caps the number of retries (default 10).
	RetryMaxCount int
	// RequestCancellation is one of the Cancel constants.
	RequestCancellation string
	// AbortController is a JavaScript expression evaluating to an AbortController, for manual
	// cancellation. It takes precedence over RequestCancellation.
	AbortController string
}

// String renders the options as a JavaScript object literal, or "" when no field is set. String
// values are quoted with Quote; Payload, AbortController and the filter are written verbatim.
func (o RequestOptions) String() string {
	var parts []string
	add := func(k, v string) { parts = append(parts, k+": "+v) }
	if o.ContentType != "" {
		add("contentType", Quote(o.ContentType))
	}
	if o.FilterSignals != nil {
		f := o.FilterSignals.String()
		if f == "" {
			f = "{include: /.*/}"
		}
		add("filterSignals", f)
	}
	if o.Selector != "" {
		add("selector", Quote(o.Selector))
	}
	if len(o.Headers) > 0 {
		keys := make([]string, 0, len(o.Headers))
		for k := range o.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		hs := make([]string, len(keys))
		for i, k := range keys {
			hs[i] = Quote(k) + ": " + Quote(o.Headers[k])
		}
		add("headers", "{"+strings.Join(hs, ", ")+"}")
	}
	if o.OpenWhenHidden != nil {
		add("openWhenHidden", strconv.FormatBool(*o.OpenWhenHidden))
	}
	if o.Payload != "" {
		add("payload", o.Payload)
	}
	if o.Retry != "" {
		add("retry", Quote(o.Retry))
	}
	if o.RetryInterval > 0 {
		add("retryInterval", strconv.FormatInt(o.RetryInterval.Milliseconds(), 10))
	}
	if o.RetryScaler != 0 {
		add("retryScaler", strconv.FormatFloat(o.RetryScaler, 'g', -1, 64))
	}
	if o.RetryMaxWait > 0 {
		add("retryMaxWait", strconv.FormatInt(o.RetryMaxWait.Milliseconds(), 10))
	}
	if o.RetryMaxCount != 0 {
		add("retryMaxCount", strconv.Itoa(o.RetryMaxCount))
	}
	switch {
	case o.AbortController != "":
		add("requestCancellation", o.AbortController)
	case o.RequestCancellation != "":
		add("requestCancellation", Quote(o.RequestCancellation))
	}
	if len(parts) == 0 {
		return ""
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// SetAll returns the expression @setAll(value, filter), setting every signal matched by filter
// (an object such as "{include: /^form\\./}", or SignalFilter.String()) to value. Both are written
// verbatim; an empty filter is omitted (all signals).
func SetAll(value, filter string) string {
	if filter == "" {
		return "@setAll(" + value + ")"
	}
	return "@setAll(" + value + ", " + filter + ")"
}

// ToggleAll returns the expression @toggleAll(filter), toggling every boolean signal matched by
// filter (written verbatim, see SignalFilter). An empty filter is omitted (all signals).
func ToggleAll(filter string) string {
	return "@toggleAll(" + filter + ")"
}

// Peek returns the expression @peek(callable), which reads signals inside callable (such as
// "() => $bar") without subscribing the surrounding expression to them.
func Peek(callable string) string { return "@peek(" + callable + ")" }

// Datastar Pro actions.

// Clipboard returns the expression @clipboard('text'), copying text to the clipboard. Pro.
// text is quoted with Quote.
func Clipboard(text string) string { return "@clipboard(" + Quote(text) + ")" }

// ClipboardBase64 returns the expression @clipboard('b64', true), copying text that is Base64
// encoded in the attribute (handy for content with quotes or code); Datastar decodes it first.
// b64 is quoted with Quote. Pro.
func ClipboardBase64(b64 string) string { return "@clipboard(" + Quote(b64) + ", true)" }

// Fit returns the expression @fit(v, oldMin, oldMax, newMin, newMax, shouldClamp, shouldRound),
// linearly mapping v from one range to another. All numbers are expressions written verbatim.
// shouldClamp and shouldRound are only written when set. Pro.
func Fit(v, oldMin, oldMax, newMin, newMax string, shouldClamp, shouldRound bool) string {
	s := "@fit(" + v + ", " + oldMin + ", " + oldMax + ", " + newMin + ", " + newMax
	switch {
	case shouldRound:
		s += ", " + strconv.FormatBool(shouldClamp) + ", true"
	case shouldClamp:
		s += ", true"
	}
	return s + ")"
}

// Intl formatter types for Intl.
const (
	IntlDateTime     = "datetime"     // Intl.DateTimeFormat
	IntlNumber       = "number"       // Intl.NumberFormat
	IntlPluralRules  = "pluralRules"  // Intl.PluralRules
	IntlRelativeTime = "relativeTime" // Intl.RelativeTimeFormat
	IntlList         = "list"         // Intl.ListFormat
	IntlDisplayNames = "displayNames" // Intl.DisplayNames
)

// Intl returns the expression @intl('type', value, options, locale), locale-aware formatting
// through the Intl namespace. typ is one of the Intl constants and is quoted; value, options (an
// object literal for the matching Intl options type) and locale (a quoted string or an array
// literal) are written verbatim. Empty options and locale are omitted; an empty options with a
// locale is written as undefined. Pro.
//
//	Intl(IntlNumber, "1000000", "{style: 'currency', currency: 'USD'}", "")
func Intl(typ, value, options, locale string) string {
	s := "@intl(" + Quote(typ) + ", " + value
	switch {
	case locale != "":
		if options == "" {
			options = "undefined"
		}
		s += ", " + options + ", " + locale
	case options != "":
		s += ", " + options
	}
	return s + ")"
}
