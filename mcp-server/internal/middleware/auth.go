package middleware

import (
	"net/http"
	"strings"

	"github.com/aibanking/mcp-server/internal/config"
)

// AuthMiddleware validates API key from header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		apiKeyHeader := config.AppConfig.Security.APIKeyHeader
		apiKey := r.Header.Get(apiKeyHeader)

		// In production, validate against database or secret manager
		// For now, accept any non-empty API key
		if apiKey == "" {
			http.Error(w, "Unauthorized: Missing API key", http.StatusUnauthorized)
			return
		}

		// Add API key to context for downstream use
		ctx := r.Context()
		// You can add API key to context here if needed

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ExtractBearerToken extracts bearer token from Authorization header
func ExtractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

