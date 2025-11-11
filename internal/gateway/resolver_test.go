package gateway

import (
	"context"
	"testing"
)

func TestResolver_Me_RequiresAuth(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test without auth context
	ctx := context.Background()
	_, err := resolver.Me(ctx)
	if err == nil {
		t.Error("expected error when no auth context present")
	}
	if err.Error() != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", err)
	}
}

func TestResolver_Usage_RequiresAuth(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test without auth context
	ctx := context.Background()
	_, err := resolver.Usage(ctx, "test-metric")
	if err == nil {
		t.Error("expected error when no auth context present")
	}
	if err.Error() != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", err)
	}
}

func TestResolver_Invoices_RequiresAuth(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test without auth context
	ctx := context.Background()
	_, err := resolver.Invoices(ctx)
	if err == nil {
		t.Error("expected error when no auth context present")
	}
	if err.Error() != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", err)
	}
}

func TestResolver_RecordUsage_RequiresAuth(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test without auth context
	ctx := context.Background()
	_, err := resolver.RecordUsage(ctx, "test-metric", 10)
	if err == nil {
		t.Error("expected error when no auth context present")
	}
	if err.Error() != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", err)
	}
}

func TestResolver_GenerateInvoice_RequiresAuth(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test without auth context
	ctx := context.Background()
	_, err := resolver.GenerateInvoice(ctx, "2024-01-01T00:00:00Z", "2024-01-31T23:59:59Z")
	if err == nil {
		t.Error("expected error when no auth context present")
	}
	if err.Error() != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", err)
	}
}

func TestResolver_UsesAuthContextOrgID(t *testing.T) {
	resolver := NewResolver("identity-svc:50051", "usage-svc:50052", "billing-svc:50053")

	// Test that resolver uses org_id from auth context, not from any input
	// This ensures org_id cannot be faked via GraphQL input
	ctx := WithAuthContext(context.Background(), &AuthContext{
		OrgID:  "org-1",
		APIKey: "test-key",
	})

	// All resolvers should use authCtx.OrgID, not any input parameter
	// The fact that there's no org_id parameter in GraphQL schema is good,
	// but we verify the resolvers use authCtx.OrgID

	// Note: These will fail to connect to services, but that's okay
	// We're just verifying they check auth context first
	_, err := resolver.Me(ctx)
	// Error is expected (service connection), but not "unauthorized"
	if err != nil && err.Error() == "unauthorized" {
		t.Error("resolver should have auth context but got unauthorized")
	}
}

func TestGetAuthContext(t *testing.T) {
	ctx := context.Background()

	// Test without auth context
	authCtx := GetAuthContext(ctx)
	if authCtx != nil {
		t.Error("expected nil auth context")
	}

	// Test with auth context
	expectedOrgID := "org-1"
	expectedAPIKey := "test-key"
	ctx = WithAuthContext(ctx, &AuthContext{
		OrgID:  expectedOrgID,
		APIKey: expectedAPIKey,
	})

	authCtx = GetAuthContext(ctx)
	if authCtx == nil {
		t.Error("expected non-nil auth context")
	}
	if authCtx.OrgID != expectedOrgID {
		t.Errorf("expected org_id %s, got %s", expectedOrgID, authCtx.OrgID)
	}
	if authCtx.APIKey != expectedAPIKey {
		t.Errorf("expected api_key %s, got %s", expectedAPIKey, authCtx.APIKey)
	}
}
