package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log after the request is done
		log.Printf(
			"Method: %s | Path: %s | Client: %s | Duration: %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			time.Since(startTime),
		)
	})
}
