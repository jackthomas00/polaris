package gateway

import (
	"net/http"
)

func AuthMiddleware(identityAddr string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			apiKey = ExtractAPIKey(authHeader)
		}

		if apiKey == "" {
			http.Error(w, "missing API key", http.StatusUnauthorized)
			return
		}

		// Validate API key
		authCtx, err := ValidateAPIKey(r.Context(), identityAddr, apiKey)
		if err != nil {
			http.Error(w, "invalid API key: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add auth context to request
		ctx := WithAuthContext(r.Context(), authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
