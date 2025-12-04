package admin

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

// InsertOrganizationMemberUser inserts a user as a member of an organization.
// If ifNotExists is true, it acts as a no-op if the user is already a member of the org.
//
// The function transactionally also adds the user to the relevant managed usergroups in the org.
// It may be called with or without holding an existing transaction.
func (s *Service) InsertOrganizationMemberUser(ctx context.Context, orgID, userID, roleID string, attributes map[string]interface{}, ifNotExists bool) error {
	ctx, tx, err := s.DB.NewTx(ctx, true)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := s.DB.InsertOrganizationMemberUser(ctx, orgID, userID, roleID, attributes, ifNotExists)
	if err != nil {
		return err
	}

	if inserted {
		err = s.DB.InsertManagedUsergroupsMemberUser(ctx, orgID, userID, roleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// InsertProjectMemberUser inserts a user as a member of a project.
// If the user is not already a member of the project's organization, it transactionally adds them as a guest of the org as well.
// It may be called with or without holding an existing transaction.
func (s *Service) InsertProjectMemberUser(ctx context.Context, orgID, projectID, userID, roleID string, attributes map[string]interface{}) error {
	guestRole, err := s.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameGuest)
	if err != nil {
		return err
	}

	ctx, tx, err := s.DB.NewTx(ctx, true)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Insert the user as a member of the project.
	err = s.DB.InsertProjectMemberUser(ctx, projectID, userID, roleID)
	if err != nil {
		return err
	}

	// All project-level members must also be org members.
	// So if the user is not already a member of the organization, add them as a guest.
	err = s.InsertOrganizationMemberUser(ctx, orgID, userID, guestRole.ID, attributes, true)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteOrganizationMemberUser deletes a user as a member of an organization.
// It transactionally also removes the user from all user groups in the org and all projects in the org.
// It may be called with or without holding an existing transaction.
func (s *Service) DeleteOrganizationMemberUser(ctx context.Context, orgID, userID string) error {
	ctx, tx, err := s.DB.NewTx(ctx, true)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = s.DB.DeleteOrganizationMemberUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// delete from all user groups of the org
	err = s.DB.DeleteUsergroupsMemberUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// delete from all projects in the org
	err = s.DB.DeleteAllProjectMemberUserForOrganization(ctx, orgID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateOrganizationMemberUserRole updates the role of a user in an organization.
// It transactionally also updates the user's membership of relevant managed usergroups in the org.
func (s *Service) UpdateOrganizationMemberUserRole(ctx context.Context, orgID, userID, roleID string) error {
	ctx, tx, err := s.DB.NewTx(ctx, true)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = s.DB.UpdateOrganizationMemberUserRole(ctx, orgID, userID, roleID)
	if err != nil {
		return err
	}

	err = s.DB.DeleteManagedUsergroupsMemberUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	err = s.DB.InsertManagedUsergroupsMemberUser(ctx, orgID, userID, roleID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CreateOrUpdateUser creates or updates a user with the given email, name, and photo URL.
// If the user doesn't exist, it creates a new user and simultaneously adds them to any orgs and projects they have been invited to.
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
			DisplayName:          name,
			PhotoURL:             photoURL,
			GithubUsername:       user.GithubUsername,
			GithubToken:          user.GithubToken,
			GithubTokenExpiresOn: user.GithubTokenExpiresOn,
			GithubRefreshToken:   user.GithubRefreshToken,
			QuotaSingleuserOrgs:  user.QuotaSingleuserOrgs,
			QuotaTrialOrgs:       user.QuotaTrialOrgs,
			PreferenceTimeZone:   user.PreferenceTimeZone,
		})
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// User does not exist. Creating a new user.

	// Get user invites if exists
	orgInvites, err := s.DB.FindOrganizationInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	projectInvites, err := s.DB.FindProjectInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	ctx, tx, err := s.DB.NewTx(ctx, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	isFirstUser, err := s.DB.CheckUsersEmpty(ctx)
	if err != nil {
		return nil, err
	}

	opts := &database.InsertUserOptions{
		Email:               email,
		DisplayName:         name,
		PhotoURL:            photoURL,
		QuotaSingleuserOrgs: deref(s.Biller.DefaultUserQuotas().SingleuserOrgs, -1),
		QuotaTrialOrgs:      deref(s.Biller.DefaultUserQuotas().TrialOrgs, -1),
		Superuser:           isFirstUser,
	}

	// Create user
	user, err = s.DB.InsertUser(ctx, opts)
	if err != nil {
		return nil, err
	}

	// handle org invites
	addedToOrgIDs := make(map[string]bool)
	addedToOrgNames := make([]string, 0)
	for _, invite := range orgInvites {
		org, err := s.DB.FindOrganization(ctx, invite.OrgID)
		if err != nil {
			return nil, err
		}
		err = s.InsertOrganizationMemberUser(ctx, invite.OrgID, user.ID, invite.OrgRoleID, nil, false)
		if err != nil {
			return nil, err
		}

		for _, usergroupID := range invite.UsergroupIDs {
			// check if the user group exists, need to check explicitly as tx is not completed yet
			exists, err := s.DB.CheckUsergroupExists(ctx, usergroupID)
			if err != nil {
				return nil, err
			}

			if !exists {
				// ignore if usergroup does not exist, might have been deleted before invite was accepted
				continue
			}

			err = s.DB.InsertUsergroupMemberUser(ctx, usergroupID, user.ID)
			if err != nil {
				return nil, err
			}
		}

		// NOTE: This cascades to deleting all project invites in the org.
		// That's alright because we already loaded the project invites to apply, but we must not reference them in the database after this.
		err = s.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}

		addedToOrgIDs[invite.OrgID] = true
		addedToOrgNames = append(addedToOrgNames, org.Name)
	}

	// check if users email domain is whitelisted for some organizations
	domain := email[strings.LastIndex(email, "@")+1:]
	organizationWhitelistedDomains, err := s.DB.FindOrganizationWhitelistedDomainsForDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, whitelist := range organizationWhitelistedDomains {
		// if user is already a member of the org then skip, prefer explicit invite to whitelist
		if _, ok := addedToOrgIDs[whitelist.OrgID]; ok {
			continue
		}
		org, err := s.DB.FindOrganization(ctx, whitelist.OrgID)
		if err != nil {
			return nil, err
		}
		err = s.InsertOrganizationMemberUser(ctx, whitelist.OrgID, user.ID, whitelist.OrgRoleID, nil, false)
		if err != nil {
			return nil, err
		}
		addedToOrgIDs[org.ID] = true
		addedToOrgNames = append(addedToOrgNames, org.Name)
	}

	// handle project invites
	addedToProjectIDs := make(map[string]bool)
	addedToProjectNames := make([]string, 0)
	for _, invite := range projectInvites {
		project, err := s.DB.FindProject(ctx, invite.ProjectID)
		if err != nil {
			return nil, err
		}
		err = s.InsertProjectMemberUser(ctx, project.OrganizationID, invite.ProjectID, user.ID, invite.ProjectRoleID, nil)
		if err != nil {
			return nil, err
		}
		addedToProjectIDs[project.ID] = true
		addedToProjectNames = append(addedToProjectNames, project.Name)

		// NOTE: Not deleting the project invite because its already been deleted by the CASCADE from the org invite delete.
		// As a sanity check, let's ensure an org invite was processed for this project invite.
		if !addedToOrgIDs[project.OrganizationID] {
			s.Logger.Error("project invite processed without a matching org invite", zap.String("project_id", project.ID))
		}
	}

	// check if users email domain is whitelisted for some projects
	projectWhitelistedDomains, err := s.DB.FindProjectWhitelistedDomainsForDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, whitelist := range projectWhitelistedDomains {
		// if user is already a member of the project then skip, prefer explicit invite to whitelist
		if _, ok := addedToProjectIDs[whitelist.ProjectID]; ok {
			continue
		}
		project, err := s.DB.FindProject(ctx, whitelist.ProjectID)
		if err != nil {
			return nil, err
		}
		err = s.InsertProjectMemberUser(ctx, project.OrganizationID, whitelist.ProjectID, user.ID, whitelist.ProjectRoleID, nil)
		if err != nil {
			return nil, err
		}
		addedToProjectIDs[project.ID] = true
		addedToProjectNames = append(addedToProjectNames, project.Name)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	s.Logger.Info("created user",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("name", user.DisplayName),
		zap.String("orgs", strings.Join(addedToOrgNames, ",")),
		zap.String("projects", strings.Join(addedToProjectNames, ",")),
	)

	return user, nil
}

// CreateOrganizationForUser creates a new organization with the given name and description, and adds the user as an admin.
func (s *Service) CreateOrganizationForUser(ctx context.Context, userID, email, orgName, displayName, description string) (*database.Organization, error) {
	viewerProjectRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameViewer)
	if err != nil {
		return nil, fmt.Errorf("failed to find viewer project role: %w", err)
	}

	txCtx, tx, err := s.DB.NewTx(ctx, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	defaultQuotas := s.Biller.DefaultQuotas()
	org, err := s.DB.InsertOrganization(txCtx, &database.InsertOrganizationOptions{
		Name:                                orgName,
		DisplayName:                         displayName,
		Description:                         description,
		LogoAssetID:                         nil,
		FaviconAssetID:                      nil,
		ThumbnailAssetID:                    nil,
		CustomDomain:                        "",
		DefaultProjectRoleID:                &viewerProjectRole.ID,
		QuotaProjects:                       deref(defaultQuotas.NumProjects, -1),
		QuotaDeployments:                    deref(defaultQuotas.NumDeployments, -1),
		QuotaSlotsTotal:                     deref(defaultQuotas.NumSlotsTotal, -1),
		QuotaSlotsPerDeployment:             deref(defaultQuotas.NumSlotsPerDeployment, -1),
		QuotaOutstandingInvites:             deref(defaultQuotas.NumOutstandingInvites, -1),
		QuotaStorageLimitBytesPerDeployment: deref(defaultQuotas.StorageLimitBytesPerDeployment, -1),
		BillingEmail:                        email,
		BillingCustomerID:                   "", // Populated later
		PaymentCustomerID:                   "", // Populated later
		CreatedByUserID:                     &userID,
	})
	if err != nil {
		return nil, err
	}

	err = s.prepareOrganization(txCtx, org.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	s.Logger.Info("created org", zap.String("name", orgName), zap.String("user_id", userID), zap.String("user_email", email))

	// raise never subscribed billing issue in sync to prevent race condition where first project is deployed before issue is raised and thus start trial job not submitted
	if s.Biller.Name() != "noop" {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypeNeverSubscribed,
			Metadata:  database.BillingIssueMetadataNeverSubscribed{},
			EventTime: org.CreatedOn,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	// Submit job to init org billing // TODO modify river client to allow job submission as part of transaction
	_, err = s.Jobs.InitOrgBilling(ctx, org.ID)
	if err != nil {
		s.Logger.Named("billing").Error("failed to submit job to init org billing", zap.String("org_id", org.ID), zap.String("org_name", orgName), zap.Error(err))
		return org, nil
	}

	return org, nil
}

func (s *Service) prepareOrganization(ctx context.Context, orgID, userID string) error {
	// Create the system-managed usergroups for the org
	err := s.DB.InsertManagedUsergroups(ctx, orgID)
	if err != nil {
		return err
	}

	// Add the user to the org as an admin.
	// This also takes care of adding them to the managed groups.
	role, err := s.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		return fmt.Errorf("failed to find admin role when preparing org: %w", err)
	}
	err = s.InsertOrganizationMemberUser(ctx, orgID, userID, role.ID, nil, false)
	if err != nil {
		return err
	}

	return nil
}

func deref[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}
