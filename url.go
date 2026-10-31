package htemel

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrUnsafeURL is wrapped by errors returned from URL when a URL is rejected.
var ErrUnsafeURL = errors.New("htemel: unsafe URL")

// safeSchemes are the absolute-URL schemes URL accepts. blob: is allowed because blob URLs are
// same-origin with the page that created them.
var safeSchemes = map[string]bool{
	"http": true, "https": true, "mailto": true, "tel": true, "sms": true, "ftp": true, "blob": true,
}

// URL reports whether s is safe to use in a URL attribute such as href, src or action, and
// returns it unchanged if so. Attribute values are HTML-escaped at render time, but schemes are
// not filtered, so untrusted input should be passed through URL (or URLOr) first.
//
//	href, err := htemel.URL(userInput)
//
// Policy:
//   - Relative references are allowed: "/a", "./a", "../a", "#x", "?q", "a/b", the empty string,
//     and protocol-relative "//host/path" (a valid relative reference that inherits the page's
//     scheme).
//   - Absolute URLs are allowed only with the scheme http, https, mailto, tel, sms, ftp or blob
//     (case-insensitive).
//   - data: URLs are allowed only for image/* media types other than image/svg+xml. SVG is
//     rejected because it can carry script. data:text/html and the like are rejected.
//   - Every other scheme (javascript:, vbscript:, ...) is rejected.
//
// Like browsers, URL ignores leading and trailing C0 control characters and spaces, and removes
// tab, CR and LF anywhere, so "  java\tscript:alert(1)" is detected as javascript:. Any other
// control character makes the URL unsafe, as does a string net/url cannot parse.
//
// Errors wrap ErrUnsafeURL.
func URL(s string) (string, error) {
	c := cleanURL(s)
	for _, r := range c {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("%w: control character in %q", ErrUnsafeURL, s)
		}
	}

	scheme, rest, ok := splitScheme(c)
	if ok {
		scheme = strings.ToLower(scheme)
		switch {
		case scheme == "data":
			if err := checkDataURL(rest); err != nil {
				return "", fmt.Errorf("%w: %v", ErrUnsafeURL, err)
			}
		case !safeSchemes[scheme]:
			return "", fmt.Errorf("%w: scheme %q not allowed", ErrUnsafeURL, scheme)
		}
	}

	if _, err := url.Parse(c); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsafeURL, err)
	}
	return s, nil
}

// MustURL is like URL but panics if s is unsafe. Use it for constants.
//
//	html.A().Href(htemel.MustURL("https://example.com"))
func MustURL(s string) string {
	u, err := URL(s)
	if err != nil {
		panic(err)
	}
	return u
}

// URLOr returns s if URL accepts it and fallback (for example "#") otherwise.
//
//	html.A().Href(htemel.URLOr(userInput, "#"))
func URLOr(s, fallback string) string {
	if u, err := URL(s); err == nil {
		return u
	}
	return fallback
}

// cleanURL mimics the browser pre-processing that matters for scheme detection.
func cleanURL(s string) string {
	s = strings.TrimFunc(s, func(r rune) bool { return r <= 0x20 })
	if strings.ContainsAny(s, "\t\r\n") {
		s = strings.NewReplacer("\t", "", "\r", "", "\n", "").Replace(s)
	}
	return s
}

// splitScheme returns the scheme of s and the remainder after the colon. ok is false when s has
// no scheme, i.e. it is a relative reference.
func splitScheme(s string) (scheme, rest string, ok bool) {
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z':
		case i > 0 && (ch >= '0' && ch <= '9' || ch == '+' || ch == '-' || ch == '.'):
		case ch == ':' && i > 0:
			return s[:i], s[i+1:], true
		default:
			return "", "", false
		}
	}
	return "", "", false
}

func checkDataURL(rest string) error {
	meta, _, found := strings.Cut(rest, ",")
	if !found {
		return errors.New("malformed data URL")
	}
	mt, _, _ := strings.Cut(meta, ";")
	mt = strings.ToLower(strings.TrimSpace(mt))
	if !strings.HasPrefix(mt, "image/") {
		return fmt.Errorf("data URL media type %q not allowed", mt)
	}
	if mt == "image/svg+xml" {
		return errors.New("data URL image/svg+xml not allowed (SVG can contain script)")
	}
	return nil
}
