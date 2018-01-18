package main

import (
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

// templ represents a single template.
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP method handles the HTTP request (load source file, compile the template and execute it)
// and writes the output to the http.ResponseWriter method.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, nil)
}

func main() {
	r := newRoom()
	// http.Handle maps the path pattern "/" to the function passed as the second argument
	// when the user hits http://localhost:8030/ the function will be executed.
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	// get the room going as a goroutine for everybody to connect to,
	// chatting operations occur in the background, allowing our main goroutine to run the web server.
	go r.run()
	// start the web server on port :8030 using the ListenAndServe method.
	if err := http.ListenAndServe(":8030", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
