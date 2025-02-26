package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	ctx := context.Background()
	svr := newTestServer(t)

	t.Run("Project members become org guests", func(t *testing.T) {
		_, c1 := newTestUser(t, svr, "")

		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: "org1"})
		require.NoError(t, err)

		r4, err := c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "proj1",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		u2, c2 := newTestUser(t, svr, "")

		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.Error(t, err)

		_, err = c1.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
			Email:        u2.Email,
			Role:         database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)

		r3, err := c2.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: r1.Organization.Name})
		require.NoError(t, err)
		require.Equal(t, r1.Organization.Name, r3.Organization.Name)

		_, err = c2.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: r1.Organization.Name, Name: r4.Project.Name})
		require.NoError(t, err)

		r5, err := c1.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{Organization: r1.Organization.Name})
		require.NoError(t, err)
		require.Len(t, r5.Members, 2)
		hasGuest := false
		for _, m := range r5.Members {
			if m.RoleName == database.OrganizationRoleNameGuest {
				hasGuest = true
			}
		}
		require.True(t, hasGuest)

		r6, err := c1.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
			Organization: r1.Organization.Name,
			Project:      r4.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r6.Members, 2)
	})

	// TODO
	// Ability to get the org

	// Ability to list org members

	// Ability to list orgs

	// Ability to add org members
	// Inability to add the same member twice
	// Remove org members
	// Remove a random user
	// Remove yourself from an org
	// Remove the last admin from an org

	// Ability to create, list, delete usergroups

	// Ability to assign roles to usergroups
	// Usergroup roles being followed
	// Ability to change role (SetOrganizationMemberUserRole)

	// Ability to invite org members + accept invite

	// Ability to whitelist a domain on org

	// Ability to whitelist a domain on project

}
