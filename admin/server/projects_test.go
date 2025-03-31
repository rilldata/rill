package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestProject(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	t.Run("Find projects by fingerprint", func(t *testing.T) {
		// Create a confounding project for another user, which should never surface in the below tests.
		_, c1 := fix.NewUser(t)
		r1, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		_, err = c1.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: r1.Organization.Name,
			Name:             "proj1",
			DirectoryName:    "foo",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Create a user with several projects
		_, c2 := fix.NewUser(t)
		o1, err := c2.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		_, err = c2.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: o1.Organization.Name,
			Name:             "proj2",
			DirectoryName:    "foo",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)
		o2, err := c2.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
		require.NoError(t, err)
		_, err = c2.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: o2.Organization.Name,
			Name:             "proj3",
			DirectoryName:    "baz",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)
		_, err = c2.CreateProject(ctx, &adminv1.CreateProjectRequest{
			OrganizationName: o2.Organization.Name,
			Name:             "proj4",
			DirectoryName:    "baz",
			ProdSlots:        1,
			SkipDeploy:       true,
		})
		require.NoError(t, err)

		// Find projects by fingerprint "foo"
		r2, err := c2.ListProjectsForFingerprint(ctx, &adminv1.ListProjectsForFingerprintRequest{DirectoryName: "foo"})
		require.NoError(t, err)
		require.Len(t, r2.Projects, 1)
		require.Equal(t, "proj2", r2.Projects[0].Name)

		// Find projects by fingerprint "baz"
		r3, err := c2.ListProjectsForFingerprint(ctx, &adminv1.ListProjectsForFingerprintRequest{DirectoryName: "baz"})
		require.NoError(t, err)
		require.Len(t, r3.Projects, 2)
		names := []string{r3.Projects[0].Name, r3.Projects[1].Name}
		require.Contains(t, names, "proj3")
		require.Contains(t, names, "proj4")
	})

}
