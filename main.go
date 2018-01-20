package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"trace"
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
	// ServeHTTP function pass the request r as the data argument to the Execute method
	// this tells the template to render itself using data that can be extracted from http.Request
	// which happens to include the host address that we need.
	t.templ.Execute(w, r)
}

func main() {
	// addr variable sets up a flag as a string that defaults to :8030
	var addr = flag.String("addr", ":8030", "The address of the application.")
	// flag.Parse parses the arguments and extracts the appropiate information.
	flag.Parse()
	r := newRoom()
	// New method creates an object that will send the output to the terminal.
	r.tracer = trace.New(os.Stdout)
	// http.Handle maps the path pattern "/chat" to the function passed as the second argument
	// when the user hits http://localhost:8030/ the function will be executed.
	// MustAuth function wraps templateHanlder, this will cause the execution to run through authHandler fisrt
	// it will run only to templateHandler if the request is authenticated.
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	// endpoint for the login page.
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	// get the room going as a goroutine for everybody to connect to,
	// chatting operations occur in the background, allowing our main goroutine to run the web server.
	go r.run()
	// log.Println outputs the address in the terminal.
	log.Println("Starting web server on", *addr)
	// start the web server on *addr (port :8030) using the ListenAndServe method.
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
