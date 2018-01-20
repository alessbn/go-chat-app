package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// authHandler stores http.Handler in the next field.
type authHandler struct {
	next http.Handler
}

// ServeHTTP method will look for a special cookie auth,
// and iy will use the Header and WriteHeader methods on http.ResponseWriter to redirect
// the user to a login page if the cookie is missing.
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")
	if err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// succes - call the next handler
	h.next.ServeHTTP(w, r)
}

// MustAuth helper function simply creates authHandler that wraps any other http.Handler.
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		log.Println("TODO handle login for", provider)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
