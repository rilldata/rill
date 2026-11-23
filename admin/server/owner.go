package server

import (
	"github.com/rilldata/rill/admin/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// owner identifies a caller that can own personal objects, such as personal files and bookmarks.
// Exactly one of userID and embedUserID is set.
type owner struct {
	userID      string
	embedUserID string
}

// ownerFromClaims resolves the owner for the caller.
// Only Rill users and embed users can own personal objects.
func ownerFromClaims(claims auth.Claims) (*owner, error) {
	switch claims.OwnerType() {
	case auth.OwnerTypeUser:
		return &owner{userID: claims.OwnerID()}, nil
	case auth.OwnerTypeEmbed:
		return &owner{embedUserID: claims.OwnerID()}, nil
	default:
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}
}

// isEmbed returns true if the owner is an embed user.
func (o *owner) isEmbed() bool {
	return o.embedUserID != ""
}

func (o *owner) userIDPtr() *string {
	if o.userID == "" {
		return nil
	}
	return &o.userID
}

func (o *owner) embedUserIDPtr() *string {
	if o.embedUserID == "" {
		return nil
	}
	return &o.embedUserID
}

// matches returns true if the owner is the one identified by the given user ID or embed user ID.
// It takes the owner columns of a database object, of which exactly one is expected to be set.
func (o *owner) matches(userID, embedUserID *string) bool {
	if o.isEmbed() {
		return embedUserID != nil && *embedUserID == o.embedUserID
	}
	return userID != nil && *userID == o.userID
}
