package auth

import (
	"context"
)

type Principal struct {
	UserID    int64
	PublicID  string
	Role      string
	Scopes    []string
	SessionID string
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok || principal.UserID <= 0 {
		return Principal{}, false
	}

	return principal, true
}
