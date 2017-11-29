package main

import (
	"log"
	"net/http"
)

func main() {
	// http.HandleFunc maps the path pattern "/" to the function passed as the second argument
	// when the user hits http://localhost:8080/ the function will be executed
	// func(w http.ResponseWriter, r *http.Request) handles HTTP requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head>
				<title>Chat</title>
			</head>
			<body>
				Let's chat!
			</body>
			</html>
		`))
	})
	// start the web server on port :8080 using the ListenAndServe method
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
