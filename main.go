package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/alessbn/go-chat-trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// set the active Avatar implementation
var avatars Avatar = TryAvatars{
	UseFileSystemAvatar,
	UseAuthAvatar,
	UseGravatar,
}

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
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	// ServeHTTP function pass the request r as the data argument to the Execute method
	// this tells the template to render itself using data that can be extracted from http.Request
	// which happens to include the host address that we need.
	t.templ.Execute(w, data)
}

func main() {
	// addr variable sets up a flag as a string that defaults to :8030
	var addr = flag.String("addr", ":8030", "The address of the application.")
	// flag.Parse parses the arguments and extracts the appropiate information.
	flag.Parse()
	// setup gomniauth
	gomniauth.SetSecurityKey("Put your auth key here.")
	gomniauth.WithProviders(
		facebook.New("key", "secret", "http://localhost:8030/auth/callback/facebook"),
		github.New("key", "secret", "http://localhost:8030/auth/callback/github"),
		google.New("key", "secret", "http://localhost:8030/auth/callback/google"),
	)
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
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	// http.StripPrefix function takes http.Handler in, modifies the path by removing the specifies prefix,
	// passes the function onto http.FileServer handler that will serve static files, provide index listings
	// and generate 404 error if it cannot find the file.
	// http.Dir allows to specify which folder we want to expose publicly.
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	http.HandleFunc("/uploader", uploaderHandler)
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))
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
