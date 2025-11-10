package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func AuthMiddleware(identityAddr string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an introspection query
		if isIntrospectionQuery(r) {
			// Allow introspection queries without authentication
			next.ServeHTTP(w, r)
			return
		}

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

func isIntrospectionQuery(r *http.Request) bool {
	// Check query parameter (for GET requests)
	if query := r.URL.Query().Get("query"); query != "" {
		return strings.Contains(query, "__schema") || strings.Contains(query, "__type")
	}

	// Check request body (for POST requests)
	if r.Body == nil {
		return false
	}

	// Read the body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		return false
	}

	// Restore the body so the GraphQL handler can read it
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// First check raw body string (most reliable, catches everything)
	bodyStr := string(bodyBytes)
	if strings.Contains(bodyStr, "IntrospectionQuery") ||
		strings.Contains(bodyStr, "__schema") ||
		strings.Contains(bodyStr, "__type") {
		return true
	}

	// Also check JSON structure for introspection
	var reqBody struct {
		Query         string `json:"query"`
		OperationName string `json:"operationName"`
	}
	if err := json.Unmarshal(bodyBytes, &reqBody); err == nil {
		// Check operationName (most reliable indicator)
		if reqBody.OperationName == "IntrospectionQuery" {
			return true
		}
		// Check query field for introspection keywords
		if strings.Contains(reqBody.Query, "__schema") || strings.Contains(reqBody.Query, "__type") {
			return true
		}
	}

	return false
}
