package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSavePersonalFileForUserAccess(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	// A project with an admin, a viewer and a deployment whose token acts on behalf of users.
	_, admin := fix.NewUser(t)
	org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	proj, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "proj1",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	viewerUser, viewer := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
		Org:     org.Organization.Name,
		Project: proj.Project.Name,
		Email:   viewerUser.Email,
		Role:    database.ProjectRoleNameViewer,
	})
	require.NoError(t, err)
	outsiderUser, _ := fix.NewUser(t)
	depl := newTestDeployment(t, fix, proj.Project.Id)

	// A deployment of another project.
	proj2, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "proj2",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	otherDepl := newTestDeployment(t, fix, proj2.Project.Id)

	// A personal canvas owned by the admin.
	dbProj, err := fix.Admin.DB.FindProject(ctx, proj.Project.Id)
	require.NoError(t, err)
	adminUserID := *dbProj.CreatedByUserID
	err = fix.Admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.Project.Id,
		Environment: "prod",
		Path:        "personal/admin-canvas.yaml",
		OwnerID:     &adminUserID,
		Data:        []byte("type: canvas\nrows: []\n"),
	})
	require.NoError(t, err)

	save := func(c *client.Client, projectID, userID, name string) error {
		_, err := c.SavePersonalFileForUser(ctx, &adminv1.SavePersonalFileForUserRequest{
			ProjectId:   projectID,
			UserId:      userID,
			Kind:        runtime.ResourceKindCanvas,
			Name:        name,
			DisplayName: "Sales",
			Yaml:        "type: canvas\nrows: []\n",
		})
		return err
	}

	t.Run("a user cannot save on behalf of another user", func(t *testing.T) {
		err := save(viewer, proj.Project.Id, adminUserID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("a deployment cannot save to another project", func(t *testing.T) {
		err := save(otherDepl, proj.Project.Id, viewerUser.ID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("the owner must be a user", func(t *testing.T) {
		err := save(depl, proj.Project.Id, "ext_8c1f0c2b", "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("the owner must have access to the project", func(t *testing.T) {
		err := save(depl, proj.Project.Id, outsiderUser.ID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("a user cannot edit another user's personal file", func(t *testing.T) {
		err := save(depl, proj.Project.Id, viewerUser.ID, "admin-canvas")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("editing a missing personal file fails", func(t *testing.T) {
		err := save(depl, proj.Project.Id, viewerUser.ID, "missing-canvas")
		requireCode(t, codes.NotFound, err)
	})

	t.Run("only canvases are supported", func(t *testing.T) {
		_, err := depl.SavePersonalFileForUser(ctx, &adminv1.SavePersonalFileForUserRequest{
			ProjectId:   proj.Project.Id,
			UserId:      viewerUser.ID,
			Kind:        runtime.ResourceKindExplore,
			DisplayName: "Sales",
			Yaml:        "type: explore\n",
		})
		requireCode(t, codes.InvalidArgument, err)
	})
}

// newTestDeployment inserts a deployment for the project, makes it the primary deployment, and returns a client authenticated with its token.
func newTestDeployment(t *testing.T, fix *testadmin.Fixture, projectID string) *client.Client {
	ctx := t.Context()
	depl, err := fix.Admin.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         projectID,
		Environment:       "prod",
		Branch:            "main",
		RuntimeHost:       "http://localhost:9091",
		RuntimeInstanceID: "test-" + randomName(),
		RuntimeAudience:   "http://localhost:8081",
		Status:            database.DeploymentStatusRunning,
		DesiredStatus:     database.DeploymentStatusRunning,
	})
	require.NoError(t, err)
	proj, err := fix.Admin.DB.FindProject(ctx, projectID)
	require.NoError(t, err)
	_, err = fix.Admin.DB.UpdateProject(ctx, projectID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		DirectoryName:        proj.DirectoryName,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GitRemote:            proj.GitRemote,
		GithubInstallationID: proj.GithubInstallationID,
		GithubRepoID:         proj.GithubRepoID,
		ManagedGitRepoID:     proj.ManagedGitRepoID,
		Provisioner:          proj.Provisioner,
		ProdVersion:          proj.ProdVersion,
		PrimaryBranch:        proj.PrimaryBranch,
		Subpath:              proj.Subpath,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		PrimaryDeploymentID:  &depl.ID,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
		OverrideDiskGB:       proj.OverrideDiskGB,
		Annotations:          proj.Annotations,
	})
	require.NoError(t, err)

	tkn, err := fix.Admin.IssueDeploymentAuthToken(ctx, depl.ID, nil)
	require.NoError(t, err)
	return fix.NewClient(t, tkn.Token().String())
}

func requireCode(t *testing.T, code codes.Code, err error) {
	t.Helper()
	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok, err)
	require.Equal(t, code, s.Code(), err)
}
