package admin

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/rilldata/rill/admin/database"
)

func (s *Service) CreateOrUpdateUser(ctx context.Context, email, name, photoURL string) (*database.User, error) {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email address %q", email)
	}

	// Find and create or update user
	user, err := s.DB.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return s.DB.CreateUser(ctx, email, name, photoURL)
		}
		return nil, err
	}

	return s.DB.UpdateUser(ctx, user.ID, name, photoURL)
}
