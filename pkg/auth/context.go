package auth

import (
	"context"
)

type AuthContext struct {
	OrgID  string
	APIKey string
}

type contextKey string

const authContextKey contextKey = "auth"

func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

func GetAuthContext(ctx context.Context) *AuthContext {
	authCtx, _ := ctx.Value(authContextKey).(*AuthContext)
	return authCtx
}
