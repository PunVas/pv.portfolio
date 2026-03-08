package server

import (
	"log"
	"net/http"
	"portfolio-server/internal/api"
	"portfolio-server/internal/data"
	"portfolio-server/internal/discord"
)

// secureHeaders middleware adds security headers to all responses
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// StartHTTP initializes and starts the HTTP server.
func StartHTTP(addr string, store *data.Store, dc *discord.Client) error {
	mux := http.NewServeMux()

	api.RegisterRoutes(mux, store, dc)

	log.Printf("[http] serving on http://%s", addr)
	return http.ListenAndServe(addr, secureHeaders(mux))
}
