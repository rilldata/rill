package server_test

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

func TestSavePersonalFileForUser(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)
	rt := newFakeRuntime(t)

	// A project with an admin, a viewer and a primary deployment whose token acts on behalf of users.
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
	depl := newTestDeployment(t, fix, proj.Project.Id, rt.host, true)

	// A deployment of the same project that isn't the primary one, and a deployment of another project.
	branchDepl := newTestDeployment(t, fix, proj.Project.Id, rt.host, false)
	proj2, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "proj2",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	otherDepl := newTestDeployment(t, fix, proj2.Project.Id, rt.host, true)

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

	save := func(c *client.Client, projectID, userID, name string) (*adminv1.SavePersonalFileForUserResponse, error) {
		return c.SavePersonalFileForUser(ctx, &adminv1.SavePersonalFileForUserRequest{
			ProjectId:   projectID,
			UserId:      userID,
			Kind:        runtime.ResourceKindCanvas,
			Name:        name,
			DisplayName: "Sales",
			// The caller tries to take over the file and share it.
			Yaml: "type: canvas\nrows: []\nannotations:\n  admin_owner_user_id: " + adminUserID + "\n  admin_shared: \"true\"\n",
		})
	}

	t.Run("a user cannot save on behalf of another user", func(t *testing.T) {
		_, err := save(viewer, proj.Project.Id, adminUserID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("a deployment cannot save to another project", func(t *testing.T) {
		_, err := save(otherDepl, proj.Project.Id, viewerUser.ID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("only the primary deployment can save", func(t *testing.T) {
		_, err := save(branchDepl, proj.Project.Id, viewerUser.ID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("the owner must be a user", func(t *testing.T) {
		_, err := save(depl, proj.Project.Id, "ext_8c1f0c2b", "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("the owner must have access to the project", func(t *testing.T) {
		_, err := save(depl, proj.Project.Id, outsiderUser.ID, "")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("a user cannot edit another user's personal file", func(t *testing.T) {
		_, err := save(depl, proj.Project.Id, viewerUser.ID, "admin-canvas")
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("editing a missing personal file fails", func(t *testing.T) {
		_, err := save(depl, proj.Project.Id, viewerUser.ID, "missing-canvas")
		requireCode(t, codes.NotFound, err)
	})

	t.Run("names outside the generated alphabet are rejected", func(t *testing.T) {
		_, err := save(depl, proj.Project.Id, viewerUser.ID, "../admin-canvas")
		requireCode(t, codes.InvalidArgument, err)
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

	t.Run("creates a private canvas owned by the user", func(t *testing.T) {
		triggers := rt.triggers.Load()
		res, err := save(depl, proj.Project.Id, viewerUser.ID, "")
		require.NoError(t, err)
		require.Regexp(t, `^sales-[0-9a-f]{8}$`, res.Name)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/canvas/%s", fix.Admin.URLs.Frontend(), org.Organization.Name, proj.Project.Name, res.Name), res.Url)
		require.Greater(t, rt.triggers.Load(), triggers, "the parser was not triggered")

		vf, err := fix.Admin.DB.FindVirtualFile(ctx, proj.Project.Id, "prod", "personal/"+res.Name+".yaml")
		require.NoError(t, err)
		require.Equal(t, viewerUser.ID, *vf.OwnerID)
		doc := parsePersonalFile(t, vf.Data)
		require.Equal(t, "canvas", doc.Type)
		require.Equal(t, "Sales", doc.DisplayName)
		require.Equal(t, viewerUser.ID, doc.Annotations["admin_owner_user_id"])
		require.NotContains(t, doc.Annotations, "admin_shared")
	})

	t.Run("editing keeps the sharing state and the owner", func(t *testing.T) {
		res, err := save(depl, proj.Project.Id, viewerUser.ID, "")
		require.NoError(t, err)
		path := "personal/" + res.Name + ".yaml"

		// The user shares it from the UI.
		err = fix.Admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID:   proj.Project.Id,
			Environment: "prod",
			Path:        path,
			OwnerID:     &viewerUser.ID,
			Data:        []byte("type: canvas\nrows: []\nannotations:\n  admin_owner_user_id: " + viewerUser.ID + "\n  admin_shared: \"true\"\n"),
		})
		require.NoError(t, err)

		// The edit can neither unshare it nor change its owner, and its display name comes from the YAML.
		_, err = depl.SavePersonalFileForUser(ctx, &adminv1.SavePersonalFileForUserRequest{
			ProjectId:   proj.Project.Id,
			UserId:      viewerUser.ID,
			Kind:        runtime.ResourceKindCanvas,
			Name:        res.Name,
			DisplayName: "Ignored on edit",
			Yaml:        "type: canvas\ndisplay_name: Sales by campus\nrows: []\nannotations:\n  admin_owner_user_id: " + adminUserID + "\n  admin_shared: \"false\"\n",
		})
		require.NoError(t, err)

		vf, err := fix.Admin.DB.FindVirtualFile(ctx, proj.Project.Id, "prod", path)
		require.NoError(t, err)
		require.Equal(t, viewerUser.ID, *vf.OwnerID)
		doc := parsePersonalFile(t, vf.Data)
		require.Equal(t, "Sales by campus", doc.DisplayName)
		require.Equal(t, viewerUser.ID, doc.Annotations["admin_owner_user_id"])
		require.Equal(t, "true", doc.Annotations["admin_shared"])
	})
}

func TestUpdateVirtualFileDataIfUnchanged(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

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

	const path = "personal/sales.yaml"
	put := func(data string) *database.VirtualFile {
		err := fix.Admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{ProjectID: proj.Project.Id, Environment: "prod", Path: path, Data: []byte(data)})
		require.NoError(t, err)
		vf, err := fix.Admin.DB.FindVirtualFile(ctx, proj.Project.Id, "prod", path)
		require.NoError(t, err)
		return vf
	}

	// Updating the version that was read succeeds.
	read := put("v1")
	require.NoError(t, fix.Admin.DB.UpdateVirtualFileDataIfUnchanged(ctx, proj.Project.Id, "prod", path, []byte("v2"), read.UpdatedOn))

	// Updating a version that changed since it was read fails, and keeps the concurrent change.
	read = put("v3")
	put("shared concurrently")
	err = fix.Admin.DB.UpdateVirtualFileDataIfUnchanged(ctx, proj.Project.Id, "prod", path, []byte("v4"), read.UpdatedOn)
	require.ErrorIs(t, err, database.ErrNotFound)
	vf, err := fix.Admin.DB.FindVirtualFile(ctx, proj.Project.Id, "prod", path)
	require.NoError(t, err)
	require.Equal(t, "shared concurrently", string(vf.Data))

	// Updating a deleted file fails instead of restoring it.
	put("v5")
	require.NoError(t, fix.Admin.DB.UpdateVirtualFileDeleted(ctx, proj.Project.Id, "prod", path))
	deleted, err := fix.Admin.DB.FindVirtualFile(ctx, proj.Project.Id, "prod", path)
	require.NoError(t, err)
	err = fix.Admin.DB.UpdateVirtualFileDataIfUnchanged(ctx, proj.Project.Id, "prod", path, []byte("v6"), deleted.UpdatedOn)
	require.ErrorIs(t, err, database.ErrNotFound)
}

// fakeRuntime is a runtime that only answers the calls the admin service makes when saving a personal file:
// it reports that no canvas exists yet, and counts parser triggers.
type fakeRuntime struct {
	runtimev1.UnimplementedRuntimeServiceServer
	host     string
	triggers atomic.Int64
}

func newFakeRuntime(t *testing.T) *fakeRuntime {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	rt := &fakeRuntime{host: "http://" + lis.Addr().String()}
	srv := grpc.NewServer()
	runtimev1.RegisterRuntimeServiceServer(srv, rt)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return rt
}

func (r *fakeRuntime) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	return nil, status.Error(codes.NotFound, "resource not found")
}

func (r *fakeRuntime) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	r.triggers.Add(1)
	return &runtimev1.CreateTriggerResponse{}, nil
}

type personalFileDoc struct {
	Type        string         `yaml:"type"`
	DisplayName string         `yaml:"display_name"`
	Annotations map[string]any `yaml:"annotations"`
}

func parsePersonalFile(t *testing.T, data []byte) personalFileDoc {
	var doc personalFileDoc
	require.NoError(t, yaml.Unmarshal(data, &doc))
	return doc
}

// newTestDeployment inserts a prod deployment for the project, optionally makes it the primary deployment, and returns a client authenticated with its token.
func newTestDeployment(t *testing.T, fix *testadmin.Fixture, projectID, runtimeHost string, primary bool) *client.Client {
	ctx := t.Context()
	depl, err := fix.Admin.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         projectID,
		Environment:       "prod",
		Branch:            "main",
		RuntimeHost:       runtimeHost,
		RuntimeInstanceID: "test-" + randomName(),
		RuntimeAudience:   "http://localhost:8081",
		Status:            database.DeploymentStatusRunning,
		DesiredStatus:     database.DeploymentStatusRunning,
	})
	require.NoError(t, err)

	if primary {
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
	}

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
