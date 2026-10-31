package handlers

import (
	"net/http"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/examples/webapp/views"
)

// GetRoot serves the home page. htemel.Handler renders into a buffer and sends a 500 (and no
// partial page) if rendering fails.
var GetRoot = htemel.Handler(func(r *http.Request) htemel.Node {
	return views.Layout(views.PageData{Title: "htemel Example WebApp", Request: r}, views.Index())
})

// GetAbout serves the about page.
var GetAbout = htemel.Handler(func(r *http.Request) htemel.Node {
	return views.Layout(views.PageData{Title: "htemel Example WebApp - About", Request: r}, views.About())
})

// GetContact serves the contact page. It uses htemel.Write directly to choose its own error page.
func GetContact(w http.ResponseWriter, r *http.Request) {
	view := views.Layout(views.PageData{Title: "htemel Example WebApp - Contact", Request: r}, views.Contact())
	if err := htemel.Write(w, r, view); err != nil {
		http.Error(w, "Failed to render view", http.StatusInternalServerError)
	}
}
