package datastar_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	nethtml "golang.org/x/net/html"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

func TestAttributes(t *testing.T) {
	tests := []struct {
		got        htemel.Attribute
		name, want string
	}{
		{datastar.Text("$x"), "data-text", "$x"},
		{datastar.Show("$x"), "data-show", "$x"},
		{datastar.Effect("$y = $x"), "data-effect", "$y = $x"},
		{datastar.Init("@get('/a')"), "data-init", "@get('/a')"},
		{datastar.Init("x", datastar.Delay(time.Second)), "data-init__delay.1000ms", "x"},
		{datastar.Signal("count", "1"), "data-signals:count", "1"},
		{datastar.Signal("user.firstName", "''"), "data-signals:user.first-name", "''"},
		{datastar.SignalsJSON("{a: 1}"), "data-signals", "{a: 1}"},
		{datastar.Computed("totalPrice", "$a*$b"), "data-computed:total-price", "$a*$b"},
		{datastar.Bind("fooBar"), "data-bind:foo-bar", ""},
		{datastar.Ref("el"), "data-ref:el", ""},
		{datastar.Indicator("loading"), "data-indicator:loading", ""},
		{datastar.JSONSignals("", datastar.Terse()), "data-json-signals__terse", ""},
		{datastar.Class("font-bold", "$b"), "data-class:font-bold", "$b"},
		{datastar.Classes("{hidden: $h}"), "data-class", "{hidden: $h}"},
		{datastar.Attr("aria-busy", "$busy"), "data-attr:aria-busy", "$busy"},
		{datastar.Attrs("{disabled: $d}"), "data-attr", "{disabled: $d}"},
		{datastar.Style("background-color", "$c"), "data-style:background-color", "$c"},
		{datastar.Styles("{color: $c}"), "data-style", "{color: $c}"},
		{datastar.Ignore(), "data-ignore", ""},
		{datastar.Ignore(datastar.Self()), "data-ignore__self", ""},
		{datastar.IgnoreMorph(), "data-ignore-morph", ""},
		{datastar.PreserveAttr("open", "class"), "data-preserve-attr", "open class"},
		{datastar.On("my-event", "x", datastar.CaseMod(datastar.CaseCamel)), "data-on:my-event__case.camel", "x"},
		{datastar.OnClick("x", datastar.Outside(), datastar.Capture(), datastar.Passive(), datastar.Stop()), "data-on:click__outside__capture__passive__stop", "x"},
		{datastar.OnSubmit("x"), "data-on:submit", "x"},
		{datastar.OnInput("x", datastar.Throttle(time.Second).NoLeading().Trailing()), "data-on:input__throttle.1000ms.noleading.trailing", "x"},
		{datastar.OnChange("x", datastar.Debounce(500*time.Millisecond).NoTrailing()), "data-on:change__debounce.500ms.notrailing", "x"},
		{datastar.OnKeyDown("x", datastar.ViewTransition()), "data-on:keydown__viewtransition", "x"},
		{datastar.OnIntersect("x", datastar.Once(), datastar.Half()), "data-on-intersect__once__half", "x"},
		{datastar.OnIntersect("x", datastar.Full()), "data-on-intersect__full", "x"},
		{datastar.OnInterval("x"), "data-on-interval", "x"},
		{datastar.OnSignalPatch("x"), "data-on-signal-patch", "x"},
		{datastar.OnSignalPatchFilter("{include: /a/}"), "data-on-signal-patch-filter", "{include: /a/}"},
		{datastar.Persist("", datastar.Session()), "data-persist__session", ""},
		{datastar.QueryString("{include: /q/}"), "data-query-string", "{include: /q/}"},
		{datastar.ReplaceURL("`/p/${$page}`"), "data-replace-url", "`/p/${$page}`"},
		{datastar.ScrollIntoView(datastar.Mod("smooth")), "data-scroll-into-view__smooth", ""},
		{datastar.CustomValidity("$ok ? '' : 'bad'"), "data-custom-validity", "$ok ? '' : 'bad'"},
		{datastar.OnRAF("x"), "data-on-raf", "x"},
		{datastar.OnResize("x"), "data-on-resize", "x"},
		{datastar.Custom("foo", "bar", "baz", datastar.Mod("m", "1", "2")), "data-foo:bar__m.1.2", "baz"},
	}
	for _, tt := range tests {
		if tt.got.Name != tt.name || tt.got.Value != tt.want || tt.got.Unsafe {
			t.Errorf("got %+v, want name %q value %q (escaped)", tt.got, tt.name, tt.want)
		}
	}
}

func TestModifierImmutable(t *testing.T) {
	base := datastar.Debounce(time.Second)
	a := base.Leading()
	b := base.NoTrailing()
	if base.String() != "__debounce.1000ms" || a.String() != "__debounce.1000ms.leading" || b.String() != "__debounce.1000ms.notrailing" {
		t.Errorf("base=%s a=%s b=%s", base, a, b)
	}
	if (datastar.Modifier{}).String() != "" {
		t.Error("zero Modifier should render empty")
	}
}

func TestQuote(t *testing.T) {
	tests := map[string]string{
		"":                         `''`,
		"/a":                       `'/a'`,
		`it's`:                     `'it\'s'`,
		`a\b`:                      `'a\\b'`,
		"a\nb\r\tc":                `'a\nb\r\tc'`,
		"\x00\x1f\x7f":             `'\x00\x1f\x7f'`,
		"\xe2\x80\xa8\xe2\x80\xa9": "'\\u2028\\u2029'",
		"héllo":                    `'héllo'`,
	}
	for in, want := range tests {
		if got := datastar.Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestActions(t *testing.T) {
	tests := map[string]string{
		datastar.Get("/a"):                    `@get('/a')`,
		datastar.Post("/a", "{x: 1}"):         `@post('/a', {x: 1})`,
		datastar.Put("/a", ""):                `@put('/a')`,
		datastar.Patch("/a?b='c'"):            `@patch('/a?b=\'c\'')`,
		datastar.Delete("/a"):                 `@delete('/a')`,
		datastar.SetAll("false", ""):          `@setAll(false)`,
		datastar.SetAll("''", "{include: 1}"): `@setAll('', {include: 1})`,
		datastar.ToggleAll(""):                `@toggleAll()`,
	}
	for got, want := range tests {
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

func TestSignalsPanicsOnUnencodable(t *testing.T) {
	if _, err := datastar.MarshalSignals(map[string]any{"c": make(chan int)}); err == nil {
		t.Fatal("MarshalSignals: expected error")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("Signals: expected panic")
		}
	}()
	datastar.Signals(func() {})
}

// TestSignalsRoundTrip renders hostile signal values and parses the HTML back, proving the
// attribute escaping is reversed by an HTML parser into the exact JSON Datastar will read.
func TestSignalsRoundTrip(t *testing.T) {
	in := map[string]any{
		"quote":  `"double" & 'single'`,
		"markup": `</div><script>alert(1)</script>`,
		"nested": map[string]any{"n": 1.5, "ok": true, "list": []any{"a", nil}},
		"crlf":   "a\r\nb",
	}
	s, err := htemel.RenderString(html.Div().Attrs(datastar.Signals(in), datastar.OnClick(datastar.Post("/it's"))))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(s, "<script>") {
		t.Fatalf("unescaped markup in %s", s)
	}

	doc, err := nethtml.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	attrs := map[string]string{}
	var walk func(*nethtml.Node)
	walk = func(n *nethtml.Node) {
		if n.Type == nethtml.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				attrs[a.Key] = a.Val
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var out map[string]any
	if err := json.Unmarshal([]byte(attrs["data-signals"]), &out); err != nil {
		t.Fatalf("data-signals %q is not JSON: %v", attrs["data-signals"], err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Errorf("round trip mismatch:\n got %#v\nwant %#v", out, in)
	}
	if got, want := attrs["data-on:click"], `@post('/it\'s')`; got != want {
		t.Errorf("data-on:click = %q, want %q", got, want)
	}
}
