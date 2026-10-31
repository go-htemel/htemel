package generators

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-htemel/spec"
)

func TestGoIdent(t *testing.T) {
	for in, want := range map[string]string{
		"a":                   "A",
		"div":                 "Div",
		"accept-charset":      "AcceptCharset",
		"http-equiv":          "HttpEquiv",
		"clipPath":            "ClipPath",
		"feGaussianBlur":      "FeGaussianBlur",
		"viewBox":             "ViewBox",
		"preserveAspectRatio": "PreserveAspectRatio",
		"stroke-width":        "StrokeWidth",
		"stroke-linecap":      "StrokeLinecap",
		"xlink:href":          "XlinkHref",
		"xml:space":           "XmlSpace",
		"xmlns:xlink":         "XmlnsXlink",
		"x-ua-compatible":     "XUaCompatible",
		"1":                   "V1",
	} {
		if got := goIdent(in); got != want {
			t.Errorf("goIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCheckNamesCollisions(t *testing.T) {
	str := func(n string) spec.Attribute { return &spec.AttributeTypeString{Name: n} }
	for name, elems := range map[string][]*spec.Element{
		"case-only tags":      {{Tag: "clipPath"}, {Tag: "clippath"}},
		"mangled attributes":  {{Tag: "path", Attributes: []spec.Attribute{str("stroke-width"), str("strokeWidth")}}},
		"built-in method":     {{Tag: "g", Attributes: []spec.Attribute{str("render")}}},
		"void built-in still": {{Tag: "br", Void: true, Attributes: []spec.Attribute{str("attr")}}},
	} {
		if err := CheckNames(elems); err == nil {
			t.Errorf("%s: expected a collision error", name)
		}
	}
	if err := CheckNames([]*spec.Element{{Tag: "br", Void: true, Attributes: []spec.Attribute{str("text")}}}); err != nil {
		t.Errorf("void element has no Text method, so a text attribute is fine: %v", err)
	}
}

// TestGenerateSVGFixture generates a small SVG spec, parses every file, checks the case-sensitive and
// mangled names, then builds the package and renders a tree with it.
func TestGenerateSVGFixture(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "svg_fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	sp := spec.Spec{}
	if err = json.Unmarshal(b, &sp); err != nil {
		t.Fatal(err)
	}

	// Generate inside the module (under testdata, so ./... never sees it) so the package can import
	// github.com/go-htemel/htemel/internal/render.
	dir, err := os.MkdirTemp("testdata", "gen-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	pkgDir := filepath.Join(dir, "svg")

	if err = Generate(&sp, pkgDir, "svg"); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	all := &strings.Builder{}
	for _, e := range sp.Elements {
		path := filepath.Join(pkgDir, strings.ToLower(e.Tag)+".go")
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected lower-cased file for <%s>: %v", e.Tag, err)
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name.Name != "svg" {
			t.Errorf("%s: package %s, want svg", path, f.Name.Name)
		}
		all.Write(src)
	}
	code := strings.Join(strings.Fields(all.String()), " ") // whitespace-insensitive (gofmt aligns)

	for _, want := range []string{
		"func ClipPath(",
		"type ClipPathElement struct",
		"func LinearGradient(",
		"func FeGaussianBlur(",
		"func (e *PathElement) StrokeWidth(s string) *PathElement",
		"func (e *PathElement) StrokeLinecap(a PathStrokeLinecap) *PathElement",
		"PathStrokeLinecapButt PathStrokeLinecap = \"butt\"",
		"func (e *SvgElement) ViewBox(",
		"func (e *SvgElement) PreserveAspectRatio(",
		"func (e *SvgElement) XmlnsXlink(",
		"func (e *LinearGradientElement) XlinkHref(",
		"func (e *GElement) XmlSpace(",
		"func (e *CircleElement) Opacity(f float64)",
		"func (e *CircleElement) Tabindex(i int)",
		"func (e *RectElement) Class(s ...string)",
		`"stroke-width"`,
		`"stroke-linecap"`,
		`"viewBox"`,
		`"preserveAspectRatio"`,
		`"xlink:href"`,
		`"xml:space"`,
		`"<clipPath"`,
		`"</clipPath>"`,
		`"<feGaussianBlur"`,
		`"<linearGradient"`,
		`render.RawText("script", text)`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("generated code is missing %s", want)
		}
	}
	for _, bad := range []string{"Clippath", "Lineargradient", "Fegaussianblur", "Viewbox", "<clippath", `"viewbox"`} {
		if strings.Contains(code, bad) {
			t.Errorf("generated code contains case-folded name %s", bad)
		}
	}

	if testing.Short() {
		return
	}

	mainDir := filepath.Join(dir, "main")
	if err = os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const mainSrc = `package main

import (
	"fmt"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/internal/generators/` + "DIR" + `/svg"
)

func main() {
	out, err := htemel.RenderString(
		svg.Svg(
			svg.ClipPath().Id("c"),
			svg.Path().D("M0 0L1 1").Stroke("currentColor").StrokeWidth("2").StrokeLinecap(svg.PathStrokeLinecapRound),
			svg.FeGaussianBlur().StdDeviation("3"),
			svg.Script().Text("if (a < b) {}"),
		).ViewBox("0 0 24 24").XmlSpace("preserve"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
`
	src := strings.Replace(mainSrc, "DIR", filepath.ToSlash(dir), 1)
	if err = os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./"+filepath.ToSlash(mainDir))
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building/running the generated package failed: %v\n%s", err, got)
	}
	want := `<svg viewBox="0 0 24 24" xml:space="preserve"><clipPath id="c"></clipPath>` +
		`<path d="M0 0L1 1" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>` +
		`<feGaussianBlur stdDeviation="3"></feGaussianBlur><script>if (a < b) {}</script></svg>`
	if string(got) != want {
		t.Errorf("rendered\n%s\nwant\n%s", got, want)
	}
}

// TestGenerateSpecFlags checks the doc comments and raw-text handling driven by the spec's
// deprecated, url and raw_text fields.
func TestGenerateSpecFlags(t *testing.T) {
	sp := spec.Spec{
		Name: "Flags",
		Elements: []*spec.Element{
			{Tag: "old", Description: "old", Deprecated: true},
			{Tag: "raw", Description: "raw", RawText: true},
			{Tag: "script", Description: "legacy spec without flag"},
			{Tag: "plain", Description: "plain", Attributes: []spec.Attribute{
				&spec.AttributeTypeString{Name: "href", URL: true},
				&spec.AttributeTypeString{Name: "title"},
				&spec.AttributeTypeNumber{Name: "size", Deprecated: true},
				&spec.AttributeTypeBool{Name: "hidden", Deprecated: true},
			}},
		},
	}
	dir := t.TempDir()
	if err := Generate(&sp, dir, "flags"); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	if old := read("old.go"); !strings.Contains(old, "// Deprecated: <old>") {
		t.Errorf("deprecated element missing Deprecated doc:\n%s", old)
	}
	if plain := read("plain.go"); strings.Contains(plain, "// Deprecated: <plain>") {
		t.Error("non-deprecated element must not carry a Deprecated doc")
	}
	for _, tag := range []string{"raw", "script"} {
		if src := read(tag + ".go"); !strings.Contains(src, "render.RawText(\""+tag+"\"") {
			t.Errorf("<%s> should be a raw text element", tag)
		}
	}
	if old := read("old.go"); strings.Contains(old, "render.RawText") {
		t.Error("<old> must not be raw text")
	}

	plain := read("plain.go")
	for _, want := range []string{
		"// Value is a URL; validate untrusted input with htemel.URL.\nfunc (e *PlainElement) Href(",
		"// Deprecated: the \"size\" attribute",
		"// Deprecated: the \"hidden\" attribute",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("plain.go missing %q:\n%s", want, plain)
		}
	}
	if i := strings.Index(plain, "func (e *PlainElement) Title("); i < 0 || strings.Contains(plain[:i][strings.LastIndex(plain[:i], "}"):], "URL") {
		t.Error("non-URL string attribute must not get the URL doc")
	}
}
