package views

import (
	. "github.com/go-htemel/htemel"
	. "github.com/go-htemel/htemel/html"
)

func Index() FlowContent {
	return PageContent("htemel Example WebApp",
		P().Text("Welcome to the htemel example webapp."),
		P().Text("Click on the links above to navigate."),
		P().Text(`
This example application showcases some of the features of htemel and how easy it is to build re-usable components.
No templating magic to learn.
Just plain old Go code.
The code we all love to read and write.
`),
	)
}
