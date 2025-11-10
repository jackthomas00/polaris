package gateway

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	identityv1 "github.com/jackthomas00/polaris/proto/identityv1"
)

type AuthContext struct {
	OrgID  string
	APIKey string
}

func ValidateAPIKey(ctx context.Context, identityAddr, apiKey string) (*AuthContext, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key")
	}

	// Connect to identity service
	conn, err := grpc.NewClient(identityAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to identity service: %w", err)
	}
	defer conn.Close()

	client := identityv1.NewIdentityClient(conn)
	resp, err := client.ValidateApiKey(ctx, &identityv1.ValidateApiKeyRequest{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid API key")
	}

	return &AuthContext{
		OrgID:  resp.OrgId,
		APIKey: apiKey,
	}, nil
}

func ExtractAPIKey(authHeader string) string {
	// Support both "Bearer <key>" and "X-API-Key: <key>" formats
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if strings.HasPrefix(authHeader, "ApiKey ") {
		return strings.TrimPrefix(authHeader, "ApiKey ")
	}
	return authHeader
}
