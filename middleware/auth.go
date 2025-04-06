package middleware

import (
	"log"
	"net/http"
)

const AUTH_TOKEN_HEADER = "X-watchman-token"

func hasAuthHeader(r *http.Request) bool {
	for k := range r.Header {
		if k == AUTH_TOKEN_HEADER {
			return true
		}
	}
	return false
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if hasAuthHeader(r) {
			v := r.Header.Get(AUTH_TOKEN_HEADER)
			log.Printf("Authenticated request, token is: %s\n", v)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}

	})
}
