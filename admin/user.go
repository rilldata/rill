package admin

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/pkg/errors"
	"github.com/rilldata/rill/admin/database"
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
		return s.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
			DisplayName:    name,
			PhotoURL:       photoURL,
			GithubUsername: user.GithubUsername,
		})
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// Get user invites if exists
	orgInvites, err := s.DB.FindOrganizationInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	projectInvites, err := s.DB.FindProjectInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// User does not exist. Creating a new user.
	user, err = s.DB.InsertUser(ctx, &database.InsertUserOptions{
		Email:       email,
		DisplayName: name,
		PhotoURL:    photoURL,
	})
	if err != nil {
		return nil, err
	}

	// handle org invites
	for _, invite := range orgInvites {
		err = s.DB.InsertOrganizationMemberUser(ctx, invite.OrgID, user.ID, invite.OrgRoleID)
		if err != nil {
			return nil, err
		}
		err = s.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}
	}

	// handle project invites
	for _, invite := range projectInvites {
		err = s.DB.InsertProjectMemberUser(ctx, invite.ProjectID, user.ID, invite.ProjectRoleID)
		if err != nil {
			return nil, err
		}
		err = s.DB.DeleteProjectInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) CreateOrganizationForUser(ctx context.Context, userID, orgName, description string) (*database.Organization, error) {
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	org, err := s.DB.InsertOrganization(ctx, &database.InsertOrganizationOptions{
		Name:        orgName,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	org, err = s.prepareOrganization(ctx, org.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (s *Service) InviteUserToOrganization(ctx context.Context, email, inviterID, orgID, roleID, orgName, roleName string) error {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid user email address %q", email)
	}

	// Create invite
	err = s.DB.InsertOrganizationInvite(ctx, email, orgID, roleID, inviterID)
	if err != nil {
		return err
	}

	// Send invitation email
	err = s.email.SendOrganizationInvite(email, "", orgName, roleName)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) InviteUserToProject(ctx context.Context, email, inviterID, projectID, roleID, projectName, roleName string) error {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid user email address %q", email)
	}

	// Create invite
	err = s.DB.InsertProjectInvite(ctx, email, projectID, roleID, inviterID)
	if err != nil {
		return err
	}

	// Send invitation email
	err = s.email.SendProjectInvite(email, "", projectName, roleName)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) prepareOrganization(ctx context.Context, orgID, userID string) (*database.Organization, error) {
	// create all user group for this org
	userGroup, err := s.DB.InsertUsergroup(ctx, &database.InsertUsergroupOptions{
		OrgID: orgID,
		Name:  "all-users",
	})
	if err != nil {
		return nil, err
	}
	// update org with all user group
	org, err := s.DB.UpdateOrganizationAllUsergroup(ctx, orgID, userGroup.ID)
	if err != nil {
		return nil, err
	}

	role, err := s.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(errors.Wrap(err, "failed to find organization admin role"))
	}

	// Add user to created org with org admin role
	err = s.DB.InsertOrganizationMemberUser(ctx, orgID, userID, role.ID)
	if err != nil {
		return nil, err
	}
	// Add user to all user group
	err = s.DB.InsertUsergroupMember(ctx, userGroup.ID, userID)
	if err != nil {
		return nil, err
	}
	return org, nil
}
