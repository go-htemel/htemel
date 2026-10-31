package tests

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/mathml"
	"github.com/go-htemel/htemel/svg"
)

// Compile-time assertions: categories, element-specific child interfaces and wildcards.
var (
	_ htemel.FlowContent             = html.Div()
	_ htemel.PalpableContent         = html.Div()
	_ htemel.PhrasingContent         = html.Span()
	_ htemel.FlowContent             = html.Span() // phrasing content is flow content
	_ htemel.InteractiveContent      = html.A()
	_ htemel.PhrasingContent         = html.A()
	_ htemel.EmbeddedContent         = html.Img()
	_ htemel.HeadingContent          = html.H1()
	_ htemel.SectioningContent       = html.Section()
	_ htemel.MetadataContent         = html.Meta()
	_ htemel.MetadataContent         = html.Title()
	_ htemel.ScriptSupportingContent = html.Script()
	_ htemel.TextContent             = htemel.Text("t")
	_ htemel.PhrasingContent         = htemel.TextUnsafe("t")

	_ html.UlChild       = html.Li()
	_ html.UlChild       = html.Script()
	_ html.UlChild       = html.Template()
	_ html.OlChild       = html.Li()
	_ html.HtmlChild     = html.Head()
	_ html.HtmlChild     = html.Body()
	_ html.TableChild    = html.Tr()
	_ html.TableChild    = html.Tbody()
	_ html.TableChild    = html.Caption()
	_ html.TrChild       = html.Td()
	_ html.TrChild       = html.Th()
	_ html.SelectChild   = html.Option()
	_ html.SelectChild   = html.Hr()
	_ html.DlChild       = html.Dt()
	_ html.DlChild       = html.Div()
	_ html.FigureChild   = html.Figcaption()
	_ html.FigureChild   = html.P()
	_ html.FigureChild   = htemel.Text("t")
	_ html.FigureChild   = svg.Svg() // through htemel.FlowMarker, from another package
	_ html.DetailsChild  = html.Summary()
	_ html.RubyChild     = html.Rt()
	_ html.RubyChild     = mathml.Math()
	_ html.AudioChild    = html.Source()
	_ html.PictureChild  = html.Img()
	_ html.ColgroupChild = html.Col()

	// svg and math are embedded, phrasing, flow and palpable content in HTML.
	_ htemel.PhrasingContent = svg.Svg()
	_ htemel.EmbeddedContent = svg.Svg()
	_ htemel.PhrasingContent = mathml.Math()
	_ svg.Content            = svg.Circle()
	_ svg.Content            = htemel.Text("t")
	_ mathml.Content         = mathml.Mi()

	// Wildcards are accepted everywhere.
	_ html.UlChild           = htemel.Group()
	_ html.HtmlChild         = htemel.Doctype()
	_ html.TrChild           = htemel.Each([]int{}, func(int) htemel.Node { return nil })
	_ htemel.TextContent     = htemel.If(true, nil)
	_ htemel.MetadataContent = htemel.Generic("x-meta", nil)
	_ svg.Content            = htemel.Unchecked(html.Div())
	_ mathml.Content         = htemel.ForEach(nil, func(int) htemel.Node { return nil })
)

// Badge is a component type of its own: embedding PhrasingMarker makes it phrasing content.
type Badge struct {
	htemel.PhrasingMarker
	label string
}

func (b Badge) Render(w io.Writer) error { return html.Span().Class("badge").Text(b.label).Render(w) }

func TestContentModelRender(t *testing.T) {
	expect(t, html.P(htemel.Text("a "), Badge{label: "new"}, html.Em().Text("b")),
		`<p>a <span class="badge">new</span><em>b</em></p>`)
	expect(t, html.Ul(html.Li().Text("1"), htemel.Each([]string{"2", "3"}, func(s string) htemel.Node {
		return html.Li().Text(s)
	})), `<ul><li>1</li><li>2</li><li>3</li></ul>`)
	expect(t, html.Table(html.Thead(html.Tr(html.Th().Text("h"))), html.Tbody(html.Tr(html.Td().Text("d")))),
		`<table><thead><tr><th>h</th></tr></thead><tbody><tr><td>d</td></tr></tbody></table>`)
	expect(t, html.Iframe().Src("/x"), `<iframe src="/x"></iframe>`)
	expect(t, html.Template(html.Tr(), html.Li().Text("x")), `<template><tr></tr><li>x</li></template>`)
	expect(t, html.Figure(html.Figcaption().Text("c"), svg.Svg()), `<figure><figcaption>c</figcaption><svg></svg></figure>`)
	expect(t, html.Div(htemel.Unchecked(nil), htemel.If(false, html.P())), `<div></div>`)
}

// TestContentModelCompileErrors builds small programs against the real packages and checks that
// content the specification does not permit is rejected by the compiler, while valid content
// compiles.
func TestContentModelCompileErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compile tests in -short mode")
	}

	// Build inside the module (under testdata, so ./... never sees it), so the programs use the
	// working tree without a go.mod of their own.
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("testdata", "cm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
		_ = os.Remove("testdata") // only succeeds if nothing else is there
	})

	build := func(name, body string) (string, error) {
		pkg := filepath.Join(dir, name)
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		src := `package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

var (
	_ = htemel.Text
	_ = svg.Svg
)

func main() {
	_ = ` + body + `
}
`
		if err := os.WriteFile(filepath.Join(pkg, "main.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("go", "vet", "./"+filepath.ToSlash(pkg)).CombinedOutput()
		return string(out), err
	}

	for name, body := range map[string]string{
		"valid_document":  `htemel.Group(htemel.Doctype(), html.Html(html.Head(html.Title().Text("t")), html.Body(html.Div(html.P(html.Span())))))`,
		"valid_list":      `html.Ul(html.Li(html.Div()), html.Script(), htemel.Group(html.Li()))`,
		"valid_svg":       `html.Div(html.Span(svg.Svg(svg.G(svg.Circle()))))`,
		"valid_text":      `html.Div().Text("x").Children(html.P().Textf("%d", 1))`,
		"valid_unchecked": `html.P(htemel.Unchecked(html.Div()))`,
	} {
		if out, err := build(name, body); err != nil {
			t.Errorf("%s: expected %s to compile:\n%s", name, body, out)
		}
	}

	for name, c := range map[string]struct{ body, want string }{
		"div_in_p":         {`html.P(html.Div())`, "does not implement htemel.PhrasingContent"},
		"div_in_ul":        {`html.Ul(html.Div())`, "missing method HTMLUlChild"},
		"text_in_ul":       {`html.Ul(htemel.Text("x"))`, "missing method HTMLUlChild"},
		"ul_text_method":   {`html.Ul().Text("x")`, "Text undefined"},
		"img_children":     {`html.Img().Children(html.Span())`, "Children undefined"},
		"iframe_children":  {`html.Iframe(html.P())`, "too many arguments"},
		"div_in_tr":        {`html.Tr(html.Div())`, "missing method HTMLTrChild"},
		"div_in_script":    {`html.Script(html.Div())`, "does not implement htemel.TextContent"},
		"li_in_div":        {`html.Div(html.Li())`, "does not implement htemel.FlowContent"},
		"body_in_div":      {`html.Div(html.Body())`, "does not implement htemel.FlowContent"},
		"div_in_head":      {`html.Head(html.Div())`, "does not implement htemel.MetadataContent"},
		"div_in_html":      {`html.Html(html.Div())`, "missing method HTMLHtmlChild"},
		"html_in_svg":      {`svg.Svg(html.Div())`, "does not implement svg.Content"},
		"svg_child_in_div": {`html.Div(svg.Circle())`, "does not implement htemel.FlowContent"},
		"node_in_div":      {`html.Div(htemel.Node(html.P()))`, "does not implement htemel.FlowContent"},
	} {
		out, err := build(name, c.body)
		if err == nil {
			t.Errorf("%s: expected %s to fail to compile", name, c.body)
			continue
		}
		if !strings.Contains(out, c.want) {
			t.Errorf("%s: compile error for %s does not mention %q:\n%s", name, c.body, c.want, out)
		}
	}
}

// Shared enum types: one type per attribute, used by every element with that attribute.
func TestSharedEnums(t *testing.T) {
	expect(t, html.Div().Popover(html.PopoverManual), `<div popover="manual"></div>`)
	expect(t, html.Span().Popover(), `<span popover></span>`)
	expect(t, html.Button().Dir(html.DirRtl).Globals(html.Globals{Translate: html.TranslateNo}), `<button dir="rtl" translate="no"></button>`)
	expect(t, html.Img().Loading(html.LoadingLazy).Crossorigin(html.CrossoriginAnonymous), `<img loading="lazy" crossorigin="anonymous">`)

	// The same type on different elements.
	v := html.PopoverAuto
	expect(t, html.Div().Popover(v), `<div popover="auto"></div>`)
	expect(t, html.Section().Popover(v), `<section popover="auto"></section>`)
	divSetter := (*html.DivElement).Popover
	spanSetter := (*html.SpanElement).Popover
	_ = func(v ...html.Popover) { divSetter(html.Div(), v...); spanSetter(html.Span(), v...) }

	// Attributes whose values differ between elements keep per-element types.
	expect(t, html.Button().Type(html.ButtonTypeSubmit), `<button type="submit"></button>`)
	expect(t, html.Input().Type(html.InputTypeEmail), `<input type="email">`)

	// svg shares enums too.
	expect(t, svg.Path().StrokeLinecap(svg.StrokeLinecapRound), `<path stroke-linecap="round"></path>`)
	expect(t, svg.Circle().StrokeLinecap(svg.StrokeLinecapRound), `<circle stroke-linecap="round"></circle>`)
}
