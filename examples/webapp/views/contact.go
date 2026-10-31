package views

import (
	. "github.com/go-htemel/htemel"
	. "github.com/go-htemel/htemel/html"
)

func Contact() FlowContent {
	return PageContent("Contact",
		P().Text("This is the contact page."),
	)
}
