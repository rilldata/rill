package auth

import (
	"context"

	"github.com/rilldata/rill/admin"
)

// Claims resolves permissions for a requester.
type Claims interface {
	Subject() string
	// TODO: Add functions for checking permissions
}

// claimsContextKey is used to set and get Claims on a request context.
type claimsContextKey struct{}

// GetClaims retrieves Claims from a request context.
// It should only be used in handlers intercepted by UnaryServerInterceptor or StreamServerInterceptor.
func GetClaims(ctx context.Context) Claims {
	claims, ok := ctx.Value(claimsContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return claims
}

// anonClaims represents claims for an unauthenticated user.
type anonClaims struct{}

func (c anonClaims) Subject() string {
	return ""
}

// tokenClaims represents claims for an auth token.
type tokenClaims struct {
	token admin.AuthToken
}

func (c *tokenClaims) Subject() string {
	return c.token.OwnerID()
}
