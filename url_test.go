package htemel_test

import (
	"errors"
	"testing"

	"github.com/go-htemel/htemel"
)

func TestURL(t *testing.T) {
	tests := []struct {
		in string
		ok bool
	}{
		{"", true},
		{"/path", true},
		{"./a", true},
		{"../a", true},
		{"#frag", true},
		{"?q=1", true},
		{"a/b:c", true},
		{"page.html", true},
		{"//evil.com", true},
		{"//evil.com/x", true},
		{"/café/日本語", true},
		{"https://example.com/a?b=c#d", true},
		{"HTTP://EXAMPLE.COM", true},
		{"mailto:a@b.c", true},
		{"tel:+123", true},
		{"sms:+123", true},
		{"ftp://host/f", true},
		{"blob:https://example.com/uuid", true},
		{"  /padded  ", true},
		{"data:image/png;base64,AAAA", true},
		{"DATA:image/gif;base64,AAAA", true},
		{"data:image/jpeg,xx", true},
		{"data:image/webp;base64,AA", true},

		{"javascript:alert(1)", false},
		{"JAVASCRIPT:alert(1)", false},
		{"JavaScript:alert(1)", false},
		{"java\tscript:alert(1)", false},
		{"java\nscript:alert(1)", false},
		{"  javascript:alert(1)", false},
		{"\x01javascript:alert(1)", false},
		{"\x00javascript:alert(1)", false},
		{"vbscript:msgbox", false},
		{"file:///etc/passwd", false},
		{"unknown:foo", false},
		{"data:text/html,<script>alert(1)</script>", false},
		{"data:text/html;base64,AAAA", false},
		{"data:image/svg+xml;base64,AAAA", false},
		{"DATA:IMAGE/SVG+XML,<svg/>", false},
		{"data:,hello", false},
		{"data:image/png", false},
		{"/a\x00b", false},
		{"/a\x7fb", false},
		{"http://[::1", false},
		{"%zz", false},
	}
	for _, tt := range tests {
		got, err := htemel.URL(tt.in)
		if tt.ok {
			if err != nil || got != tt.in {
				t.Errorf("URL(%q) = %q, %v; want unchanged, nil", tt.in, got, err)
			}
			continue
		}
		if err == nil || !errors.Is(err, htemel.ErrUnsafeURL) || got != "" {
			t.Errorf("URL(%q) = %q, %v; want ErrUnsafeURL", tt.in, got, err)
		}
	}
}

func TestMustURL(t *testing.T) {
	if got := htemel.MustURL("/ok"); got != "/ok" {
		t.Errorf("got %q", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	htemel.MustURL("javascript:alert(1)")
}

func TestURLOr(t *testing.T) {
	if got := htemel.URLOr("https://a.b", "#"); got != "https://a.b" {
		t.Errorf("got %q", got)
	}
	if got := htemel.URLOr("javascript:x", "#"); got != "#" {
		t.Errorf("got %q", got)
	}
}
