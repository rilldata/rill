package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestListRoles(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	t.Run("Listing roles", func(t *testing.T) {
		_, c := fix.NewUser(t)
		res, err := c.ListRoles(ctx, &adminv1.ListRolesRequest{})
		require.NoError(t, err)

		require.Len(t, res.OrganizationRoles, 4)
		require.Len(t, res.ProjectRoles, 3)

		var orgAdminRole *adminv1.OrganizationRole
		for _, r := range res.OrganizationRoles {
			if r.Name == "admin" {
				orgAdminRole = r
				break
			}
		}
		require.NotNil(t, orgAdminRole)
		require.NotEmpty(t, orgAdminRole.Id)
		require.True(t, orgAdminRole.Permissions.Admin)
		require.False(t, orgAdminRole.Permissions.Guest)
		require.True(t, orgAdminRole.Permissions.ManageOrg)

		var projViewerRole *adminv1.ProjectRole
		for _, r := range res.ProjectRoles {
			if r.Name == "viewer" {
				projViewerRole = r
				break
			}
		}
		require.NotNil(t, projViewerRole)
		require.NotEmpty(t, projViewerRole.Id)
		require.False(t, projViewerRole.Permissions.Admin)
		require.False(t, projViewerRole.Permissions.ManageProject)
		require.True(t, projViewerRole.Permissions.ReadProd)
	})

	t.Run("Default project role", func(t *testing.T) {
		// Create a user and an org
		_, c := fix.NewUser(t)
		org, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)

		// Create a project and check the autogroup:members group has the default viewer role
		proj1, err := c.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		groups1, err := c.ListProjectMemberUsergroups(ctx, &adminv1.ListProjectMemberUsergroupsRequest{
			Org:     org.Organization.Name,
			Project: proj1.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, groups1.Members, 1)
		require.Equal(t, database.UsergroupNameAutogroupMembers, groups1.Members[0].GroupName)
		require.Equal(t, "viewer", groups1.Members[0].RoleName)
		require.True(t, groups1.Members[0].GroupManaged)

		// Update the default role to editor
		_, err = c.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
			Org:                org.Organization.Name,
			DefaultProjectRole: toPtr("editor"),
		})
		require.NoError(t, err)

		// Create another project and check the autogroup:members group now has the editor role
		proj2, err := c.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj2",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		groups2, err := c.ListProjectMemberUsergroups(ctx, &adminv1.ListProjectMemberUsergroupsRequest{
			Org:     org.Organization.Name,
			Project: proj2.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, groups2.Members, 1)
		require.Equal(t, database.UsergroupNameAutogroupMembers, groups2.Members[0].GroupName)
		require.Equal(t, "editor", groups2.Members[0].RoleName)
		require.True(t, groups2.Members[0].GroupManaged)

		// Update the default role to none
		_, err = c.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
			Org:                org.Organization.Name,
			DefaultProjectRole: toPtr(""),
		})
		require.NoError(t, err)

		// Create another project and check the autogroup:members group now has no role
		proj3, err := c.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        org.Organization.Name,
			Project:    "proj3",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		groups3, err := c.ListProjectMemberUsergroups(ctx, &adminv1.ListProjectMemberUsergroupsRequest{
			Org:     org.Organization.Name,
			Project: proj3.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, groups3.Members, 0)
	})
}

func toPtr[T any](v T) *T {
	return &v
}
