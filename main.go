package main

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

// templ represents a single template
type templateHandler struct {
	filename string
	templ    *template.Template
}

// ServeHTTP method handles the HTTP request (load source file, compile the template and execute it)
// and writes the output to the http.ResponseWriter method
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.templ == nil {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	}
	t.templ.Execute(w, nil)
}

func main() {
	// http.Handle maps the path pattern "/" to the function passed as the second argument
	// when the user hits http://localhost:8080/ the function will be executed
	http.Handle("/", (&templateHandler{filename: "chat.html"}))
	// start the web server on port :8080 using the ListenAndServe method
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
