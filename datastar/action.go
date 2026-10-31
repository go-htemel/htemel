package datastar

import (
	"strconv"
	"strings"
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
// argument; Datastar takes one options object, such as "{openWhenHidden: true}". Empty opts are
// skipped.
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

// SetAll returns the expression @setAll(value, filter), setting every signal matched by filter
// (an object such as "{include: /^form\\./}") to value. Both are written verbatim; an empty filter
// is omitted (all signals).
func SetAll(value, filter string) string {
	if filter == "" {
		return "@setAll(" + value + ")"
	}
	return "@setAll(" + value + ", " + filter + ")"
}

// ToggleAll returns the expression @toggleAll(filter), toggling every boolean signal matched by
// filter. An empty filter is omitted (all signals).
func ToggleAll(filter string) string {
	return "@toggleAll(" + filter + ")"
}
