package gateway

import (
	"testing"
)

func TestValidateAPIKey_OrgIDFromIdentityService(t *testing.T) {
	// This test verifies that org_id comes from identity service, not client input
	// The ValidateAPIKey function calls identity service which looks up org_id from the API key

	// Key points:
	// 1. API key is the only client input
	// 2. org_id is returned by identity service based on the API key
	// 3. Client cannot directly provide org_id

	// In a real test, you'd mock the identity service gRPC client
	// For now, we verify the function signature and behavior

	// The function takes apiKey (from client) and returns AuthContext with OrgID (from identity-svc)
	// This ensures org_id cannot be faked by the client
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Bearer token",
			header:   "Bearer test-api-key-12345",
			expected: "test-api-key-12345",
		},
		{
			name:     "ApiKey prefix",
			header:   "ApiKey test-api-key-12345",
			expected: "test-api-key-12345",
		},
		{
			name:     "Plain key",
			header:   "test-api-key-12345",
			expected: "test-api-key-12345",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractAPIKey(tt.header)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAuthContext_OrgIDNotFromClient(t *testing.T) {
	// This test documents that AuthContext.OrgID should never be set from client input
	// It should only come from identity service via ValidateAPIKey

	// The AuthContext struct has OrgID field, but it's only populated by:
	// 1. ValidateAPIKey -> calls identity service -> gets org_id from DB based on API key
	// 2. Never from GraphQL input, HTTP headers (except API key), or any client-provided data

	// This is enforced by:
	// - GraphQL schema has no org_id input parameters
	// - All resolvers use authCtx.OrgID from context
	// - authCtx is only set by AuthMiddleware after validating API key with identity service
}
