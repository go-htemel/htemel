package main

import (
	"net/http"

	"github.com/go-htemel/htemel/examples/webapp/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", handlers.GetRoot)
	mux.Handle("/about", handlers.GetAbout)
	mux.HandleFunc("/contact", handlers.GetContact)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
