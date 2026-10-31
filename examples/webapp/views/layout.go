package views

import (
	"net/http"

	. "github.com/go-htemel/htemel"
	. "github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

type PageData struct {
	Title   string
	Request *http.Request
}

// Layout wraps page content in the document shell. Children are flow content, since they are
// placed in a <div> inside <body>.
func Layout(data PageData, children ...FlowContent) Node {
	var path string
	if data.Request != nil {
		path = data.Request.URL.Path
	}

	layout := Group(
		Doctype(),
		Html().Class("bg-zinc-50 dark:bg-zinc-950").Children(
			Head().Children(
				Meta().Charset("utf-8"),
				Meta().Name("viewport").Content("width=device-width, initial-scale=1"),
				Script().Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"),
				Title().Text(data.Title),
			),
			Body().
				Class("text-zinc-900 dark:text-zinc-100").
				Children(
					Navigation(path),
					Div().Class().
						Children(children...),
				),
		),
	)

	return layout
}

type navLink struct {
	href   string
	text   string
	active bool
}

func Navigation(currentPage string) FlowContent {
	links := []navLink{
		{"/about", "About", false},
		{"/contact", "Contact", false},
	}

	return Nav().
		Class("bg-zinc-900 border-b border-zinc-800").
		Children(
			Div().
				Class("max-w-7xl mx-auto px-4 sm:px-6 lg:px-8").
				Children(
					Div().
						Class("flex justify-between h-16 items-center").
						Children(
							Div().
								Class("flex shrink-0").
								Children(
									A().Href("/").Class("flex items-center gap-2 text-white").Children(
										LogoIcon(),
										Span().Text("ExampleApp"),
									),
								),
							Div().
								Class("hidden md:flex items-center space-x-8").
								Children(
									Each(links, func(link navLink) Node {
										link.active = currentPage == link.href
										return NavigationLink(link)
									}),
								),
						),
				),
		)
}

// LogoIcon is a small inline SVG mark. It uses currentColor, so it takes the text color of the link.
func LogoIcon() PhrasingContent {
	return svg.Svg().
		Class("size-6 text-red-500").
		ViewBox("0 0 24 24").
		Attr("aria-hidden", "true").
		Children(
			svg.Path().
				D("M12 2 2 7l10 5 10-5-10-5Z").
				Fill("none").
				Stroke("currentColor").
				StrokeWidth("2").
				StrokeLinejoin(svg.StrokeLinejoinRound),
			svg.Path().
				D("M2 17l10 5 10-5M2 12l10 5 10-5").
				Fill("none").
				Stroke("currentColor").
				StrokeWidth("2").
				StrokeLinejoin(svg.StrokeLinejoinRound),
		)
}

func NavigationLink(link navLink) Node {
	return A().
		Href(link.href).
		Class("hover:text-white px-3 py-5 text-sm font-medium transition-colors border-b-2 hover:border-red-500/70").
		ToggleClass("text-gray-300 border-transparent", !link.active).
		ToggleClass("text-white border-red-500/70", link.active).
		Text(link.text)
}

func PageContent(name string, children ...FlowContent) FlowContent {
	return Div().
		Class("max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-8 space-y-4").
		Children(
			H1().
				Class("text-3xl font-bold").
				Text(name),
		).
		Children(children...)

}
