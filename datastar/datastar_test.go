package datastar_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	nethtml "golang.org/x/net/html"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// attributeCases lists every attribute helper (fn is its Go name) with the attribute it must emit.
var attributeCases = []struct {
	fn         string
	got        htemel.Attribute
	name, want string
}{
	{"Text", datastar.Text("$x"), "data-text", "$x"},
	{"Show", datastar.Show("$x"), "data-show", "$x"},
	{"Effect", datastar.Effect("$y = $x"), "data-effect", "$y = $x"},
	{"Init", datastar.Init("@get('/a')"), "data-init", "@get('/a')"},
	{"Init", datastar.Init("x", datastar.Delay(time.Second)), "data-init__delay.1000ms", "x"},
	{"Signals", mustSignals(map[string]any{"a": 1}, datastar.IfMissing()), "data-signals__ifmissing", `{"a":1}`},
	{"SignalsJSON", datastar.SignalsJSON("{a: 1}"), "data-signals", "{a: 1}"},
	{"Signal", datastar.Signal("count", "1"), "data-signals:count", "1"},
	{"Signal", datastar.Signal("user.firstName", "''"), "data-signals:user.first-name", "''"},
	{"Signal", datastar.Signal("x", "1", datastar.CaseMod(datastar.CaseSnake)), "data-signals:x__case.snake", "1"},
	{"Computed", datastar.Computed("totalPrice", "$a*$b"), "data-computed:total-price", "$a*$b"},
	{"Computeds", datastar.Computeds("{t: () => $a}"), "data-computed", "{t: () => $a}"},
	{"Bind", datastar.Bind("fooBar"), "data-bind:foo-bar", ""},
	{"Bind", datastar.Bind("isChecked", datastar.Prop("checked"), datastar.Event("input", "change")), "data-bind:is-checked__prop.checked__event.input.change", ""},
	{"Ref", datastar.Ref("el"), "data-ref:el", ""},
	{"Indicator", datastar.Indicator("loading"), "data-indicator:loading", ""},
	{"JSONSignals", datastar.JSONSignals("", datastar.Terse()), "data-json-signals__terse", ""},
	{"JSONSignals", datastar.JSONSignals("{include: /a/}"), "data-json-signals", "{include: /a/}"},
	{"Class", datastar.Class("font-bold", "$b"), "data-class:font-bold", "$b"},
	{"Class", datastar.Class("myClass", "$b", datastar.CaseMod(datastar.CaseCamel)), "data-class:myClass__case.camel", "$b"},
	{"Classes", datastar.Classes("{hidden: $h}"), "data-class", "{hidden: $h}"},
	{"Attr", datastar.Attr("aria-busy", "$busy"), "data-attr:aria-busy", "$busy"},
	{"Attrs", datastar.Attrs("{disabled: $d}"), "data-attr", "{disabled: $d}"},
	{"Style", datastar.Style("background-color", "$c"), "data-style:background-color", "$c"},
	{"Styles", datastar.Styles("{color: $c}"), "data-style", "{color: $c}"},
	{"Ignore", datastar.Ignore(), "data-ignore", ""},
	{"Ignore", datastar.Ignore(datastar.Self()), "data-ignore__self", ""},
	{"IgnoreMorph", datastar.IgnoreMorph(), "data-ignore-morph", ""},
	{"PreserveAttr", datastar.PreserveAttr("open", "class"), "data-preserve-attr", "open class"},
	{"On", datastar.On("my-event", "x", datastar.CaseMod(datastar.CaseCamel)), "data-on:my-event__case.camel", "x"},
	{"On", datastar.On("keydown", "x", datastar.Document(), datastar.Delay(time.Second)), "data-on:keydown__document__delay.1000ms", "x"},
	{"OnClick", datastar.OnClick("x", datastar.Outside(), datastar.Capture(), datastar.Passive(), datastar.Stop(), datastar.Window(), datastar.Prevent(), datastar.Once()), "data-on:click__outside__capture__passive__stop__window__prevent__once", "x"},
	{"OnSubmit", datastar.OnSubmit("x"), "data-on:submit", "x"},
	{"OnInput", datastar.OnInput("x", datastar.Throttle(time.Second).NoLeading().Trailing()), "data-on:input__throttle.1000ms.noleading.trailing", "x"},
	{"OnChange", datastar.OnChange("x", datastar.Debounce(500*time.Millisecond).NoTrailing()), "data-on:change__debounce.500ms.notrailing", "x"},
	{"OnKeyDown", datastar.OnKeyDown("x", datastar.ViewTransition()), "data-on:keydown__viewtransition", "x"},
	{"OnIntersect", datastar.OnIntersect("x", datastar.Once(), datastar.Half()), "data-on-intersect__once__half", "x"},
	{"OnIntersect", datastar.OnIntersect("x", datastar.Full(), datastar.Exit()), "data-on-intersect__full__exit", "x"},
	{"OnIntersect", datastar.OnIntersect("x", datastar.Threshold(25)), "data-on-intersect__threshold.25", "x"},
	{"OnInterval", datastar.OnInterval("x", datastar.Duration(500*time.Millisecond).Leading()), "data-on-interval__duration.500ms.leading", "x"},
	{"OnSignalPatch", datastar.OnSignalPatch("x", datastar.Debounce(time.Second).Leading()), "data-on-signal-patch__debounce.1000ms.leading", "x"},
	{"OnSignalPatchFilter", datastar.OnSignalPatchFilter("{include: /a/}"), "data-on-signal-patch-filter", "{include: /a/}"},
	// Pro.
	{"Animate", datastar.Animate("x"), "data-animate", "x"},
	{"MatchMedia", datastar.MatchMedia("isDark", "prefers-color-scheme: dark"), "data-match-media:is-dark", "'prefers-color-scheme: dark'"},
	{"ViewTransitionName", datastar.ViewTransitionName("$foo"), "data-view-transition", "$foo"},
	{"Persist", datastar.Persist("", "", datastar.Session()), "data-persist__session", ""},
	{"Persist", datastar.Persist("mykey", "{include: /a/}", datastar.Session()), "data-persist:mykey__session", "{include: /a/}"},
	{"QueryString", datastar.QueryString("{include: /q/}", datastar.Filter(), datastar.History()), "data-query-string__filter__history", "{include: /q/}"},
	{"ReplaceURL", datastar.ReplaceURL("`/p/${$page}`"), "data-replace-url", "`/p/${$page}`"},
	{"ScrollIntoView", datastar.ScrollIntoView(datastar.Smooth(), datastar.VCenter(), datastar.HNearest(), datastar.Focus()), "data-scroll-into-view__smooth__vcenter__hnearest__focus", ""},
	{"CustomValidity", datastar.CustomValidity("$ok ? '' : 'bad'"), "data-custom-validity", "$ok ? '' : 'bad'"},
	{"OnRAF", datastar.OnRAF("x", datastar.Throttle(10*time.Millisecond)), "data-on-raf__throttle.10ms", "x"},
	{"OnResize", datastar.OnResize("x", datastar.Debounce(10*time.Millisecond)), "data-on-resize__debounce.10ms", "x"},
	{"Custom", datastar.Custom("foo", "bar", "baz", datastar.Mod("m", "1", "2")), "data-foo:bar__m.1.2", "baz"},
}

func mustSignals(v any, mods ...datastar.Modifier) htemel.Attribute {
	return datastar.Signals(v, mods...)
}

func TestAttributes(t *testing.T) {
	for _, tt := range attributeCases {
		if tt.got.Name != tt.name || tt.got.Value != tt.want || tt.got.Unsafe {
			t.Errorf("%s: got %+v, want name %q value %q (escaped)", tt.fn, tt.got, tt.name, tt.want)
		}
	}
}

// TestEveryHelperCovered fails when an exported function returning htemel.Attribute is missing
// from attributeCases, so new helpers cannot ship without an asserted attribute name.
func TestEveryHelperCovered(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	covered := map[string]bool{}
	for _, c := range attributeCases {
		covered[c.fn] = true
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Recv != nil || !fd.Name.IsExported() || fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
					continue
				}
				if sel, ok := fd.Type.Results.List[0].Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "Attribute" && !covered[fd.Name.Name] {
					t.Errorf("helper %s has no case in attributeCases", fd.Name.Name)
				}
			}
		}
	}
}

// TestMatchesReference compares the attribute names the helpers emit with the list extracted from
// the Datastar v1.0.4 reference (testdata/v1.0.4-attributes.txt): every documented attribute,
// free or Pro, needs a helper, and every helper must emit a documented attribute.
func TestMatchesReference(t *testing.T) {
	raw, err := os.ReadFile("testdata/v1.0.4-attributes.txt")
	if err != nil {
		t.Fatal(err)
	}
	documented := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			documented[line] = true
		}
	}
	if len(documented) < 30 {
		t.Fatalf("reference list has only %d attributes", len(documented))
	}
	emitted := map[string]bool{}
	for _, c := range attributeCases {
		if c.fn == "Custom" {
			continue
		}
		base := c.got.Name
		if i := strings.IndexAny(base, ":"); i >= 0 {
			base = base[:i]
		}
		if i := strings.Index(base, "__"); i >= 0 {
			base = base[:i]
		}
		emitted[base] = true
		if !documented[base] {
			t.Errorf("%s emits %q, which is not in the Datastar v1.0.4 reference", c.fn, base)
		}
	}
	for name := range documented {
		if !emitted[name] {
			t.Errorf("documented attribute %s has no helper", name)
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
		datastar.Query("/a"):                  `@query('/a')`,
		datastar.SetAll("false", ""):          `@setAll(false)`,
		datastar.SetAll("''", "{include: 1}"): `@setAll('', {include: 1})`,
		datastar.ToggleAll(""):                `@toggleAll()`,
		datastar.ToggleAll(datastar.SignalFilter{Include: `^is`}.String()): `@toggleAll({include: /^is/})`,
		datastar.Peek("() => $bar"):                                        `@peek(() => $bar)`,
		datastar.Clipboard("it's"):                                         `@clipboard('it\'s')`,
		datastar.ClipboardBase64("SGk="):                                   `@clipboard('SGk=', true)`,
		datastar.Fit("$v", "0", "100", "0", "255", false, false):           `@fit($v, 0, 100, 0, 255)`,
		datastar.Fit("$v", "0", "100", "0", "1", true, false):              `@fit($v, 0, 100, 0, 1, true)`,
		datastar.Fit("$v", "0", "100", "0", "1", false, true):              `@fit($v, 0, 100, 0, 1, false, true)`,
		datastar.Fit("$v", "0", "100", "0", "1", true, true):               `@fit($v, 0, 100, 0, 1, true, true)`,
		datastar.Intl(datastar.IntlNumber, "5", "", ""):                    `@intl('number', 5)`,
		datastar.Intl(datastar.IntlNumber, "5", "{style: 'percent'}", ""):  `@intl('number', 5, {style: 'percent'})`,
		datastar.Intl(datastar.IntlDateTime, "d", "", "'de-AT'"):           `@intl('datetime', d, undefined, 'de-AT')`,
		datastar.Intl(datastar.IntlList, "x", "{}", "['en', 'de']"):        `@intl('list', x, {}, ['en', 'de'])`,
	}
	for got, want := range tests {
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

func TestRequestOptions(t *testing.T) {
	yes, no := true, false
	tests := []struct {
		o    datastar.RequestOptions
		want string
	}{
		{datastar.RequestOptions{}, ""},
		{datastar.RequestOptions{ContentType: datastar.ContentForm, Selector: "#f"}, `{contentType: 'form', selector: '#f'}`},
		{datastar.RequestOptions{FilterSignals: &datastar.SignalFilter{Include: `^foo\.`, Exclude: `_temp$`}}, `{filterSignals: {include: /^foo\./, exclude: /_temp$/}}`},
		{datastar.RequestOptions{FilterSignals: &datastar.SignalFilter{}}, `{filterSignals: {include: /.*/}}`},
		{datastar.RequestOptions{Headers: map[string]string{"X-B": "it's", "X-A": "1"}}, `{headers: {'X-A': '1', 'X-B': 'it\'s'}}`},
		{datastar.RequestOptions{OpenWhenHidden: &yes}, `{openWhenHidden: true}`},
		{datastar.RequestOptions{OpenWhenHidden: &no, Payload: "{a: $a}"}, `{openWhenHidden: false, payload: {a: $a}}`},
		{datastar.RequestOptions{Retry: datastar.RetryError, RetryInterval: 2 * time.Second, RetryScaler: 1.5, RetryMaxWait: time.Minute, RetryMaxCount: 3},
			`{retry: 'error', retryInterval: 2000, retryScaler: 1.5, retryMaxWait: 60000, retryMaxCount: 3}`},
		{datastar.RequestOptions{RequestCancellation: datastar.CancelDisabled}, `{requestCancellation: 'disabled'}`},
		{datastar.RequestOptions{RequestCancellation: datastar.CancelAuto, AbortController: "$controller"}, `{requestCancellation: $controller}`},
	}
	for _, tt := range tests {
		if got := tt.o.String(); got != tt.want {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
	if got, want := datastar.Post("/s", datastar.RequestOptions{Retry: datastar.RetryNever}.String()), `@post('/s', {retry: 'never'})`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestModifierArgs(t *testing.T) {
	if got := datastar.Event("input", "change").String(); got != "__event.input.change" {
		t.Error(got)
	}
	if got := datastar.Threshold(75).String(); got != "__threshold.75" {
		t.Error(got)
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
