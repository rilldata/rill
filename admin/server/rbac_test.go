package server

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	ctx := context.Background()
	svr := newTestServer(t)

	t.Run("Adding org and project members", func(t *testing.T) {
		// Create users
		u1, c1 := newTestUser(t, svr, "")
		u2, c2 := newTestUser(t, svr, "")

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

	// Public projects

	// Correct tracking of members in all-users, all-guests, all-members

	// Ability to invite non-existing users to orgs/projects + accept invite

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
