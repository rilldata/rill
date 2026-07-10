package server_test

import (
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestProjectVariables(t *testing.T) {
	fix := testadmin.New(t)

	t.Run("Set, get and unset variables as a user", func(t *testing.T) {
		u1, c1 := fix.NewUser(t)

		r1, err := c1.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Set a shared variable, a prod-only variable, and a dev-only variable.
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:       r1.Organization.Name,
			Project:   r2.Project.Name,
			Variables: map[string]string{"FOO": "shared"},
		})
		require.NoError(t, err)
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "prod",
			Variables:   map[string]string{"BAR": "prod-value"},
		})
		require.NoError(t, err)
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "dev",
			Variables:   map[string]string{"BAZ": "dev-value"},
		})
		require.NoError(t, err)

		// Get shared variables only.
		r3, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r3.Variables, 1)
		require.Equal(t, "FOO", r3.Variables[0].Name)
		require.Equal(t, "shared", r3.Variables[0].Value)
		require.Equal(t, "", r3.Variables[0].Environment)
		require.Equal(t, u1.ID, r3.Variables[0].UpdatedByUserId)

		// Get shared and prod variables.
		r4, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:         r1.Organization.Name,
			Project:     r2.Project.Name,
			Environment: "prod",
		})
		require.NoError(t, err)
		require.Len(t, r4.Variables, 2)
		require.Equal(t, "FOO", r4.Variables[0].Name)
		require.Equal(t, "shared", r4.Variables[0].Value)
		require.Equal(t, "BAR", r4.Variables[1].Name)
		require.Equal(t, "prod-value", r4.Variables[1].Value)

		// Get all variables across environments.
		r5, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:                r1.Organization.Name,
			Project:            r2.Project.Name,
			ForAllEnvironments: true,
		})
		require.NoError(t, err)
		require.Len(t, r5.Variables, 3)
		require.Equal(t, "FOO", r5.Variables[0].Name)
		require.Equal(t, "BAR", r5.Variables[1].Name)
		require.Equal(t, "BAZ", r5.Variables[2].Name)

		// Unset a variable.
		_, err = c1.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:            r1.Organization.Name,
			Project:        r2.Project.Name,
			UnsetVariables: []string{"FOO"},
		})
		require.NoError(t, err)

		// Check that the variable is gone.
		r6, err := c1.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:                r1.Organization.Name,
			Project:            r2.Project.Name,
			ForAllEnvironments: true,
		})
		require.NoError(t, err)
		require.Len(t, r6.Variables, 2)
		require.Equal(t, "BAR", r6.Variables[0].Name)
		require.Equal(t, "BAZ", r6.Variables[1].Name)
	})

	t.Run("Set and get variables using a service token", func(t *testing.T) {
		_, c1 := fix.NewUser(t)

		// Create org and project
		r1, err := c1.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		r2, err := c1.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
			Org:        r1.Organization.Name,
			Project:    "proj1",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)

		// Create a service with project admin role and issue an auth token for it.
		r3, err := c1.CreateService(t.Context(), &adminv1.CreateServiceRequest{
			Name:            "service1",
			Org:             r1.Organization.Name,
			Project:         r2.Project.Name,
			ProjectRoleName: database.ProjectRoleNameAdmin,
		})
		require.NoError(t, err)
		r4, err := c1.IssueServiceAuthToken(t.Context(), &adminv1.IssueServiceAuthTokenRequest{
			Org:         r1.Organization.Name,
			ServiceName: r3.Service.Name,
		})
		require.NoError(t, err)
		c2 := fix.NewClient(t, r4.Token)

		// Set a variable using the service token.
		_, err = c2.UpdateProjectVariables(t.Context(), &adminv1.UpdateProjectVariablesRequest{
			Org:       r1.Organization.Name,
			Project:   r2.Project.Name,
			Variables: map[string]string{"FOO": "bar"},
		})
		require.NoError(t, err)

		// Read it back using the service token.
		r5, err := c2.GetProjectVariables(t.Context(), &adminv1.GetProjectVariablesRequest{
			Org:     r1.Organization.Name,
			Project: r2.Project.Name,
		})
		require.NoError(t, err)
		require.Len(t, r5.Variables, 1)
		require.Equal(t, "FOO", r5.Variables[0].Name)
		require.Equal(t, "bar", r5.Variables[0].Value)
		require.Equal(t, "", r5.Variables[0].UpdatedByUserId)
	})
}
