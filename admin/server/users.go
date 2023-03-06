package server

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetCurrentUser(ctx context.Context, req *adminv1.GetCurrentUserRequest) (*adminv1.GetCurrentUserResponse, error) {
	claims := auth.GetClaims(ctx)

	// Return an empty result if not authenticated.
	ent, ok := claims.OwnerEntity()
	if !ok {
		return &adminv1.GetCurrentUserResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if ent != database.EntityUser {
		return nil, fmt.Errorf("not authenticated as a user")
	}

	// Owner is a user
	u, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	return &adminv1.GetCurrentUserResponse{
		User: userToPB(u),
	}, nil
}

func userToPB(u *database.User) *adminv1.User {
	return &adminv1.User{
		Id:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		PhotoUrl:    u.PhotoURL,
		CreatedOn:   timestamppb.New(u.CreatedOn),
		UpdatedOn:   timestamppb.New(u.UpdatedOn),
	}
}
