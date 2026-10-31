package tests

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// deepTree builds a tree depth levels deep, where the innermost div holds leaves children.
func deepTree(depth, leaves int) htemel.Node {
	inner := html.Div().Class("leaf-parent")
	for i := 0; i < leaves; i++ {
		inner.Children(html.Span().Class("leaf").Attr("data-i", strconv.Itoa(i)).Text("leaf <" + strconv.Itoa(i) + ">"))
	}
	var n htemel.FlowContent = inner
	for d := 1; d < depth; d++ {
		n = html.Div().Class("level").Id("l"+strconv.Itoa(d)).Children(n, html.P().Text("sibling & text"))
	}
	return n
}

func realisticPage() htemel.Node {
	return html.Html(
		html.Head(
			html.Title().Text("Benchmark Test"),
			html.Meta().Charset("utf-8"),
		),
		html.Body(
			html.Div().Class("container").With(func(d *html.DivElement) {
				for i := 0; i < 100; i++ {
					d.Children(
						html.Div().Class("item").
							Children(
								html.H2().Text("Item "+strconv.Itoa(i)),
								html.P().Text("Some description..."),
							),
					)
				}
			}),
		),
	)
}

// BenchmarkDeepNesting renders 8 nested levels with 20 children at the leaves.
func BenchmarkDeepNesting(b *testing.B) {
	page := deepTree(8, 20)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = page.Render(buf)
	}
}

// devNull returns an unbuffered *os.File, where every Write is a syscall, as with an
// unbuffered http.ResponseWriter or os.Stdout.
func devNull(b *testing.B) io.Writer {
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = f.Close() })
	return f
}

// BenchmarkUnbufferedFileDirect calls .Render straight into an unbuffered file.
func BenchmarkUnbufferedFileDirect(b *testing.B) {
	page := realisticPage()
	w := devNull(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = page.Render(w)
	}
}

// BenchmarkUnbufferedFileHtemelRender renders the same page into the same unbuffered file through
// htemel.Render, which buffers and flushes once.
func BenchmarkUnbufferedFileHtemelRender(b *testing.B) {
	page := realisticPage()
	w := devNull(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = htemel.Render(w, page)
	}
}

// BenchmarkRenderString measures htemel.RenderString on the realistic page.
func BenchmarkRenderString(b *testing.B) {
	page := realisticPage()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = htemel.RenderString(page)
	}
}
