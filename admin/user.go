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

	// We create an initial org with a name derived from the user's info
	err = s.createOrgForUser(ctx, email, name)
	if err != nil {
		s.logger.Ctx(ctx).Error("failed to create organization for user", zap.String("user.id", user.ID), zap.Error(err))
		// continuing, since user was created successfully
	}

	return user, nil
}

func (s *Service) createOrgForUser(ctx context.Context, email, name string) error {
	// Start a tx for creating org and adding the user
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	orgNameSeeds := nameseeds.ForUser(email, name)

	_, err = s.DB.CreateOrganizationFromSeeds(ctx, orgNameSeeds, name)
	if err != nil {
		return err
	}

	// TODO: Add user to created org

	return tx.Commit()
}
