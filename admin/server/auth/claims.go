package auth

import (
	"context"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
)

// Claims resolves permissions for a requester.
type Claims interface {
	OwnerEntity() (database.Entity, bool)
	OwnerID() string
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

func (c anonClaims) OwnerEntity() (database.Entity, bool) {
	return "", false
}

func (c anonClaims) OwnerID() string {
	return ""
}

// authTokenClaims represents claims for an admin.AuthToken.
type authTokenClaims struct {
	token admin.AuthToken
}

func (c *authTokenClaims) OwnerEntity() (database.Entity, bool) {
	return c.token.OwnerType(), true
}

func (c *authTokenClaims) OwnerID() string {
	return c.token.OwnerID()
}
