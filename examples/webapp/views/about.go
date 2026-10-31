package views

import (
	. "github.com/go-htemel/htemel"
	. "github.com/go-htemel/htemel/html"
)

func About() FlowContent {
	return PageContent("About",
		P().Text("This is the about page."),
	)
}
