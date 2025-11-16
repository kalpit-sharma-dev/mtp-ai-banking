package middleware

import (
	"net/http"

	"github.com/aibanking/banking-integrations/internal/config"
)

// AuthMiddleware validates API key from header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health check and OPTIONS (preflight) requests without auth
		if r.URL.Path == "/health" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		apiKeyHeader := config.AppConfig.Security.APIKeyHeader
		apiKey := r.Header.Get(apiKeyHeader)

		if apiKey == "" {
			http.Error(w, "Unauthorized: Missing API key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

