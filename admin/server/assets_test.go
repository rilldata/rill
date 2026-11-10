package server_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestPublicProjectArchive(t *testing.T) {
	fix := testadmin.New(t)
	owner, ownerClient := fix.NewUser(t)

	org, err := ownerClient.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)

	// Create a private deploy archive and a public project that uses it
	assetID := uuid.New().String()
	_, err = fix.Admin.DB.InsertAsset(t.Context(), assetID, org.Organization.Id, fmt.Sprintf("gs://bucket/deploy/%s.tar.gz", assetID), owner.ID, false)
	require.NoError(t, err)
	proj, err := ownerClient.CreateProject(t.Context(), &adminv1.CreateProjectRequest{
		Org:            org.Organization.Name,
		Project:        "proj",
		Public:         true,
		ArchiveAssetId: assetID,
		SkipDeploy:     true,
	})
	require.NoError(t, err)
	require.Equal(t, assetID, proj.Project.ArchiveAssetId)

	// An outside user with a valid session
	outsider, err := fix.Admin.CreateOrUpdateUser(t.Context(), "outsider@example.com", "Outsider", "")
	require.NoError(t, err)
	outsiderToken, err := fix.Admin.IssueUserAuthToken(t.Context(), outsider.ID, database.AuthClientIDRillWeb, "Test session", nil, nil, false)
	require.NoError(t, err)
	outsiderClient := fix.NewClient(t, outsiderToken.Token().String())
	anonClient := fix.NewClient(t, "")

	t.Run("archive ID is only returned to project managers", func(t *testing.T) {
		res, err := ownerClient.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: org.Organization.Name, Project: "proj"})
		require.NoError(t, err)
		require.Equal(t, assetID, res.Project.ArchiveAssetId)

		res, err = outsiderClient.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: org.Organization.Name, Project: "proj"})
		require.NoError(t, err)
		require.Empty(t, res.Project.ArchiveAssetId)

		res, err = anonClient.GetProject(t.Context(), &adminv1.GetProjectRequest{Org: org.Organization.Name, Project: "proj"})
		require.NoError(t, err)
		require.Empty(t, res.Project.ArchiveAssetId)

		list, err := anonClient.ListProjectsForOrganization(t.Context(), &adminv1.ListProjectsForOrganizationRequest{Org: org.Organization.Name})
		require.NoError(t, err)
		require.Len(t, list.Projects, 1)
		require.Empty(t, list.Projects[0].ArchiveAssetId)
	})

	t.Run("archive cannot be downloaded by visitors of the public project", func(t *testing.T) {
		for name, token := range map[string]string{"anonymous": "", "outsider": outsiderToken.Token().String()} {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("%s/v1/assets/%s/download", fix.ExternalURL(), assetID), http.NoBody)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			res.Body.Close()
			require.Equal(t, http.StatusForbidden, res.StatusCode, name)
		}
	})
}
