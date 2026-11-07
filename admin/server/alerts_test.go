package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCreateAlertValidation(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	// Create an admin user with an org and a project without a deployment.
	adminUser, admin := fix.NewUser(t)
	org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	proj, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "proj1",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)

	// Add project members with non-admin roles.
	editorUser, editor := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{Org: org.Organization.Name, Project: proj.Project.Name, Email: editorUser.Email, Role: database.ProjectRoleNameEditor})
	require.NoError(t, err)
	viewerUser, viewer := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{Org: org.Organization.Name, Project: proj.Project.Name, Email: viewerUser.Email, Role: database.ProjectRoleNameViewer})
	require.NoError(t, err)

	// Add an org admin that is not a member of the project.
	orgAdminUser, _ := fix.NewUser(t)
	_, err = admin.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{Org: org.Organization.Name, Email: orgAdminUser.Email, Role: database.OrganizationRoleNameAdmin})
	require.NoError(t, err)

	resourceStatusOpts := func(recipients ...string) *adminv1.AlertOptions {
		props, err := structpb.NewStruct(map[string]any{"where_error": true})
		require.NoError(t, err)
		return &adminv1.AlertOptions{
			DisplayName:        "Resource errors",
			RefreshCron:        "*/10 * * * *",
			Resolver:           "resource_status",
			ResolverProperties: props,
			EmailRecipients:    recipients,
		}
	}

	requireCode := func(t *testing.T, err error, code codes.Code, contains string) {
		require.Error(t, err)
		require.Equal(t, code, status.Code(err), err.Error())
		require.Contains(t, err.Error(), contains)
	}

	t.Run("resolvers other than resource_status are rejected", func(t *testing.T) {
		props, err := structpb.NewStruct(map[string]any{"sql": "SELECT * FROM m1"})
		require.NoError(t, err)
		_, err = viewer.CreateAlert(ctx, &adminv1.CreateAlertRequest{
			Org:     org.Organization.Name,
			Project: proj.Project.Name,
			Options: &adminv1.AlertOptions{
				DisplayName:        "SQL alert",
				Resolver:           "sql",
				ResolverProperties: props,
				EmailRecipients:    []string{viewerUser.Email},
			},
		})
		requireCode(t, err, codes.InvalidArgument, `unsupported alert resolver "sql"`)
	})

	t.Run("only project admins can create resource status alerts", func(t *testing.T) {
		for _, c := range []*client.Client{viewer, editor} {
			_, err := c.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: org.Organization.Name, Project: proj.Project.Name, Options: resourceStatusOpts()})
			requireCode(t, err, codes.PermissionDenied, "resource status alerts")
		}
	})

	t.Run("resource status alerts require a refresh schedule", func(t *testing.T) {
		opts := resourceStatusOpts(adminUser.Email)
		opts.RefreshCron = ""
		_, err := admin.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: org.Organization.Name, Project: proj.Project.Name, Options: opts})
		requireCode(t, err, codes.InvalidArgument, "refresh schedule")
	})

	t.Run("email recipients must be project admins", func(t *testing.T) {
		for _, email := range []string{editorUser.Email, viewerUser.Email, "unknown@example.com"} {
			_, err := admin.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: org.Organization.Name, Project: proj.Project.Name, Options: resourceStatusOpts(adminUser.Email, email)})
			requireCode(t, err, codes.InvalidArgument, email)
		}
	})

	t.Run("project and org admins are valid recipients", func(t *testing.T) {
		// Validation passes, so the request fails on the missing deployment.
		_, err := admin.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: org.Organization.Name, Project: proj.Project.Name, Options: resourceStatusOpts(adminUser.Email, orgAdminUser.Email)})
		requireCode(t, err, codes.FailedPrecondition, "production deployment")
	})
}
