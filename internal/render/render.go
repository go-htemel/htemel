// Package render holds the shared rendering helpers used by the generated element packages.
package render

import (
	"strings"

	"golang.org/x/net/html"
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

// WriteTo writes all attributes, each preceded by a space, in insertion order.
func (a *Attrs) WriteTo(sb *strings.Builder) {
	for _, at := range a.list {
		WriteAttr(sb, at.key, at.value, at.raw)
	}
}

// WriteAttr writes a single attribute preceded by a space. Empty values render bare.
func WriteAttr(sb *strings.Builder, key, value string, raw bool) {
	sb.WriteByte(' ')
	sb.WriteString(key)
	if value == "" {
		return
	}
	sb.WriteString(`="`)
	if raw {
		sb.WriteString(value)
	} else {
		sb.WriteString(html.EscapeString(value))
	}
	sb.WriteByte('"')
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
