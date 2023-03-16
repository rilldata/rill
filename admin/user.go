package admin

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/nameseeds"
	"go.uber.org/zap"
)

func (s *Service) CreateOrUpdateUser(ctx context.Context, email, name, photoURL string) (*database.User, error) {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email address %q", email)
	}

	// Update user if exists
	user, err := s.DB.FindUserByEmail(ctx, email)
	if err == nil {
		return s.DB.UpdateUser(ctx, user.ID, name, photoURL)
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// User does not exist. Creating a new user.
	user, err = s.DB.CreateUser(ctx, email, name, photoURL)
	if err != nil {
		return nil, err
	}

	// Create a default organization
	orgNameSeeds := nameseeds.ForUser(email, name)
	_, err = s.DB.CreateOrganizationFromSeeds(ctx, orgNameSeeds, name)
	if err != nil {
		s.logger.Error("failed to create organization for user", zap.Strings("seeds", orgNameSeeds), zap.String("user.id", user.ID), zap.Error(err))
		// continuing, since user was created successfully
	}

	// TODO: Add user to created org

	return user, nil
}
