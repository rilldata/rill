package server

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	ctx := context.Background()
	svr := newTestServer(t)

	t.Run("Adding org and project members", func(t *testing.T) {
		// Create users
		u1, c1 := newTestUser(t, svr)
		u2, c2 := newTestUser(t, svr)

		// Create org and project
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r4, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Check u2 cannot get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.Error(t, err)
		r5, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r5.Organizations, 0)

		// Add u2 to the project
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
			Email:        u2.Email,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Check u2 can get and list the org and project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.NoError(t, err)
		r7, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r7.Organizations, 1)
		r8, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r8.Projects, 1)

		// Check u2 is a guest in the org
		r9, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r9.Members, 2)
		hasGuest := false
		for _, m := range r9.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check u2 is an admin in the project
		r10, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r10.Members, 2)
		for _, m := range r10.Members {
			require.Equal(t, database.ProjectRoleNameAdmin, m.RoleName)
		}

		// Check we can't add u2 to the org (since they are already in it as a guest)
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.Error(t, err)

		// Check we can change u2's role in the org
		_, err = c1.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)

		// Remove u2 from the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u2.Email,
		})
		require.NoError(t, err)

		// Check they are removed from the list of project members as well
		r11, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{Organization: r1.Organization.Name, Project: r4.Project.Name})
		require.NoError(t, err)
		require.Len(t, r11.Members, 1)
		require.Equal(t, u1.Email, r11.Members[0].UserEmail)

		// Check they can't get or list the org or project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.Error(t, err)
		r12, err := c2.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Len(t, r12.Organizations, 0)

		// Check the last admin can't leave the org
		_, err = c1.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Organization: r1.Organization.Name,
			Email:        u1.Email,
		})
		require.Error(t, err)
		_, err = c1.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{Organization: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Visibility of public projects", func(t *testing.T) {
		// Create users
		_, c1 := newTestUser(t, svr)
		_, c2 := newTestUser(t, svr)

		// Create org and public and private projects
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "public",
			ProdSlots:        1,
			SkipDeploy:       true,
			Public:           true,
		})
		require.NoError(t, err)
		r3, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "private",
			ProdSlots:        1,
			SkipDeploy:       true,
			Public:           false,
		})
		require.NoError(t, err)

		// Check u2 can access the org and the public project
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r2.Project.Name})
		require.NoError(t, err)

		// Check u2 can't access or list the private project
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r3.Project.Name})
		require.Error(t, err)
		r4, err := c2.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r4.Projects, 1)
		require.Equal(t, r2.Project.Name, r4.Projects[0].Name)

		// Delete the public project
		_, err = c1.DeleteProject(ctx, &adminv1.DeleteProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             r2.Project.Name,
		})
		require.NoError(t, err)

		// Check u2 can't access the org any more
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.Error(t, err)
	})

	t.Run("Inviting users who have not signed up yet", func(t *testing.T) {
		// Create admin user with org, project and group
		_, c1 := newTestUser(t, svr)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)
		group1, err := c1.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Organization: org1.Organization.Name,
			Name:         "group1",
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the org, project and group
		userEmail := randomName() + "@example.com"
		_, err = c1.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Organization: org1.Organization.Name,
			Email:        userEmail,
			Role:         database.OrganizationRoleNameViewer,
		})
		require.NoError(t, err)
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Email:        userEmail,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		_, err = c1.AddUsergroupMemberUser(ctx, &adminv1.AddUsergroupMemberUserRequest{
			Organization: org1.Organization.Name,
			Usergroup:    group1.Usergroup.GroupName,
			Email:        userEmail,
		})
		require.NoError(t, err)

		// Check that two emails were sent (org and project addition)
		sender := svr.admin.Email.Sender.(*email.TestSender)
		var count int
		for _, email := range sender.Emails {
			if email.ToEmail == userEmail {
				count++
			}
		}
		require.Equal(t, 2, count)

		// Check that we can list the invites
		orgInvites, err := c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 1)
		require.Equal(t, userEmail, orgInvites.Invites[0].Email)
		require.Equal(t, database.OrganizationRoleNameViewer, orgInvites.Invites[0].Role)
		projInvites, err := c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 1)
		require.Equal(t, userEmail, projInvites.Invites[0].Email)
		require.Equal(t, database.ProjectRoleNameAdmin, projInvites.Invites[0].Role)

		// Create the user and check they can access the org and project, and check they are in the list of members
		_, c2 := newTestUserWithEmail(t, svr, userEmail)
		_, err = c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: org1.Organization.Name})
		require.NoError(t, err)
		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: org1.Organization.Name, Name: proj1.Project.Name})
		require.NoError(t, err)
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		projMembers, err := c2.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projMembers.Members, 2)
		groupMembers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    group1.Usergroup.GroupName,
		})
		require.NoError(t, err)
		require.Len(t, groupMembers.Members, 1)

		// Check that the invites were deleted
		orgInvites, err = c1.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgInvites.Invites, 0)
		projInvites, err = c1.ListProjectInvites(ctx, &adminv1.ListProjectInvitesRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, projInvites.Invites, 0)

		// Check that the user was added to the all-users group
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    "all-users",
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
	})

	t.Run("Inviting project users who haven't signed up yet become org guests", func(t *testing.T) {
		// Create admin user with org and project
		_, c1 := newTestUser(t, svr)
		org1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		proj1, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: org1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Invite a user that doesn't exist to the project
		userEmail := randomName() + "@example.com"
		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: org1.Organization.Name,
			Project:      proj1.Project.Name,
			Email:        userEmail,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		// Sign up the user
		_, _ = newTestUserWithEmail(t, svr, userEmail)

		// Check that the user is a guest in the org
		orgMembers, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: org1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, orgMembers.Members, 2)
		hasGuest := false
		for _, m := range orgMembers.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		// Check that the user is in the all-users group
		allUsers, err := c1.ListUsergroupMemberUsers(ctx, &adminv1.ListUsergroupMemberUsersRequest{
			Organization: org1.Organization.Name,
			Usergroup:    "all-users",
		})
		require.NoError(t, err)
		require.Len(t, allUsers.Members, 2)
	})

	// Correct tracking of members in all-users, all-guests, all-members

	// Ability to whitelist a domain on org
	// Ability to whitelist a domain on project

	// Ability to create, list, delete usergroups
	// Ability to assign roles to usergroups
	// Usergroup roles being followed
}

func randomName() string {
	id := randomBytes(16)
	return "test_" + hex.EncodeToString(id)
}
