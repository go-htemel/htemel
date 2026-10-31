// Package render holds the shared rendering helpers used by the generated element packages.
package render

import (
	"io"
	"slices"
	"strings"
)

// attr is a single attribute. Values are stored unescaped; escaping happens at render time
// unless raw is set.
type attr struct {
	key   string
	value string
	raw   bool
}

// Attrs is an insertion-ordered set of HTML attributes.
// The first Set of a key fixes its position; later Sets of the same key update the value in place.
// The zero value is ready to use.
type Attrs struct {
	list []attr
}

func (a *Attrs) index(key string) int {
	for i := range a.list {
		if a.list[i].key == key {
			return i
		}
	}
	return -1
}

func (a *Attrs) set(key, value string, raw bool) {
	if i := a.index(key); i >= 0 {
		a.list[i].value = value
		a.list[i].raw = raw
		return
	}
	a.list = append(a.list, attr{key: key, value: value, raw: raw})
}

// Set stores an attribute whose value will be HTML-escaped when rendered.
// An empty value renders as a bare attribute (equivalent to key="").
func (a *Attrs) Set(key, value string) { a.set(key, value, false) }

// SetUnsafe stores an attribute whose value is rendered verbatim, without escaping.
func (a *Attrs) SetUnsafe(key, value string) { a.set(key, value, true) }

// SetBool renders key as a bare boolean attribute when on is true and removes it when false.
func (a *Attrs) SetBool(key string, on bool) {
	if on {
		a.set(key, "", false)
		return
	}
	a.Remove(key)
}

// Get returns the (unescaped) value stored for key.
func (a *Attrs) Get(key string) (string, bool) {
	if i := a.index(key); i >= 0 {
		return a.list[i].value, true
	}
	return "", false
}

// Remove deletes key if present.
func (a *Attrs) Remove(key string) {
	if i := a.index(key); i >= 0 {
		a.list = append(a.list[:i], a.list[i+1:]...)
	}
}

// Len reports the number of attributes.
func (a *Attrs) Len() int { return len(a.list) }

// WriteTo writes all attributes, each preceded by a space, in insertion order, directly to w.
// It implements io.WriterTo.
func (a *Attrs) WriteTo(w io.Writer) (int64, error) {
	var total int64
	for i := range a.list {
		at := &a.list[i]
		n, err := WriteAttr(w, at.key, at.value, at.raw)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// WriteAttr writes a single attribute preceded by a space directly to w. Empty values render bare.
// Unless raw is set, the value is escaped exactly as html.EscapeString would, without allocating.
func WriteAttr(w io.Writer, key, value string, raw bool) (int, error) {
	total, err := io.WriteString(w, " ")
	if err != nil {
		return total, err
	}
	n, err := io.WriteString(w, key)
	total += n
	if err != nil || value == "" {
		return total, err
	}
	n, err = io.WriteString(w, `="`)
	total += n
	if err != nil {
		return total, err
	}
	if raw {
		n, err = io.WriteString(w, value)
	} else {
		n, err = WriteEscaped(w, value)
	}
	total += n
	if err != nil {
		return total, err
	}
	n, err = io.WriteString(w, `"`)
	return total + n, err
}

// escapedChars is the set escaped by golang.org/x/net/html.EscapeString (note: it includes '\r').
const escapedChars = "&'<>\"\r"

// WriteEscaped writes s to w with exactly the escaping of html.EscapeString from
// golang.org/x/net/html, but streams the unescaped runs and the replacement entities straight to w
// instead of building a new string. Values with nothing to escape are written in one call.
func WriteEscaped(w io.Writer, s string) (int, error) {
	i := strings.IndexAny(s, escapedChars)
	if i < 0 {
		return io.WriteString(w, s)
	}
	total := 0
	for i >= 0 {
		n, err := io.WriteString(w, s[:i])
		total += n
		if err != nil {
			return total, err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			esc = "&#39;" // matches x/net/html (shorter than &apos;)
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;" // matches x/net/html (shorter than &quot;)
		case '\r':
			esc = "&#13;"
		}
		n, err = io.WriteString(w, esc)
		total += n
		if err != nil {
			return total, err
		}
		s = s[i+1:]
		i = strings.IndexAny(s, escapedChars)
	}
	n, err := io.WriteString(w, s)
	return total + n, err
}

// RawText prepares text for a raw text element (script, style). The content is not HTML-escaped;
// the only transformation is that any case-insensitive occurrence of "</tag" is rewritten to
// "<\/tag" so the content cannot terminate the element early.
func RawText(tag, text string) string {
	needle := "</" + tag
	n := len(needle)
	var sb strings.Builder
	last := 0
	for i := 0; i+n <= len(text); i++ {
		if text[i] == '<' && text[i+1] == '/' && strings.EqualFold(text[i:i+n], needle) {
			if last == 0 && sb.Len() == 0 {
				sb.Grow(len(text) + 1)
			}
			sb.WriteString(text[last:i])
			sb.WriteString(`<\/`)
			sb.WriteString(text[i+2 : i+n])
			last = i + n
			i += n - 1
		}
	}
	if last == 0 {
		return text
	}
	sb.WriteString(text[last:])
	return sb.String()
}

// AppendClasses returns current with classes appended. Each element of classes may hold several
// space-separated class names. A name already present (in current or earlier in classes) is
// skipped, so order is preserved and no appended name repeats. Whitespace is normalised to single
// spaces.
func AppendClasses(current string, classes []string) string {
	names := strings.Fields(current)
	for _, c := range classes {
		for _, name := range strings.Fields(c) {
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
	}
	return strings.Join(names, " ")
}
