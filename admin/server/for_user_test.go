package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestAlertsAndReportsForUser tests creating and deleting alerts and reports for a user through a runtime's admin service, as the AI tools do.
// The admin service writes them like the dialogs do, and the runtime parses and reconciles them.
func TestAlertsAndReportsForUser(t *testing.T) {
	fix := testadmin.NewWithOptionalRuntime(t, true)
	ctx := t.Context()

	// The first user becomes a superuser, so the second one owns the org and project to avoid confounding effects.
	_, _ = fix.NewUser(t)
	ownerUser, owner := fix.NewUser(t)
	org, err := owner.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	project, err := owner.CreateProject(ctx, &adminv1.CreateProjectRequest{Org: org.Organization.Name, Project: "for-user", ProdSlots: 1, SkipDeploy: true})
	require.NoError(t, err)
	viewerUser, viewer := fix.NewUser(t)
	_, err = owner.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{Org: org.Organization.Name, Project: project.Project.Name, Email: viewerUser.Email, Role: database.ProjectRoleNameViewer})
	require.NoError(t, err)
	outsiderUser, _ := fix.NewUser(t)

	// Create a runtime instance whose admin service is this admin service, acting as the project's primary deployment.
	// The instance is not provisioned by the admin service (which requires deploying a project), so it doesn't sync the project's virtual files:
	// mirrorVirtualFiles copies them to its repo instead.
	depl, err := owner.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{Org: org.Organization.Name, Project: project.Project.Name, Environment: "prod"})
	require.NoError(t, err)
	proj, err := fix.Admin.DB.FindProject(ctx, project.Project.Id)
	require.NoError(t, err)
	require.Equal(t, depl.Deployment.Id, *proj.PrimaryDeploymentID)
	tkn, err := fix.Admin.IssueDeploymentAuthToken(ctx, depl.Deployment.Id, nil)
	require.NoError(t, err)

	repoDir := t.TempDir()
	for path, data := range map[string]string{
		"rill.yaml":        "",
		"models/bar.sql":   `SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, 'Denmark' AS country, 1 AS revenue`,
		"metrics/mv1.yaml": "version: 1\ntype: metrics_view\nmodel: bar\ntimeseries: __time\ndimensions:\n- column: country\nmeasures:\n- name: revenue\n  expression: sum(revenue)\n",
		"explores/e1.yaml": "type: explore\nmetrics_view: mv1\n",
	} {
		abs := filepath.Join(repoDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}
	instanceID := randomName()
	_, err = fix.Admin.DB.InsertProvisionerResource(ctx, &database.InsertProvisionerResourceOptions{
		ID:           uuid.NewString(),
		DeploymentID: depl.Deployment.Id,
		Type:         string(provisioner.ResourceTypeRuntime),
		Status:       database.ProvisionerResourceStatusOK,
		Provisioner:  "static",
		Config:       map[string]any{"host": fix.RuntimeURL(), "audience": fix.RuntimeURL()},
	})
	require.NoError(t, err)
	_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
		RuntimeHost:       fix.RuntimeURL(),
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   fix.RuntimeURL(),
		Status:            database.DeploymentStatusRunning,
		StatusMessage:     "Running",
	})
	require.NoError(t, err)
	inst := &drivers.Instance{
		ID:               instanceID,
		Environment:      "prod",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		AdminConnector:   "admin",
		Connectors: []*runtimev1.Connector{
			{Type: "file", Name: "repo", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": repoDir}))},
			{Type: "duckdb", Name: "duckdb", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": ":memory:", "mode": "readwrite"}))},
			{Type: "sqlite", Name: "catalog", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s-catalog?mode=memory&cache=shared", t.Name())}))},
			{Type: "admin", Name: "admin", Config: testruntime.Must(structpb.NewStruct(map[string]any{"admin_url": fix.ExternalURL(), "access_token": tkn.Token().String(), "project_id": proj.ID}))},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}
	require.NoError(t, fix.Runtime.CreateInstance(ctx, inst))
	ctrl, err := fix.Runtime.Controller(ctx, inst.ID)
	require.NoError(t, err)
	require.NoError(t, ctrl.WaitUntilIdle(ctx, false))
	testruntime.RequireReconcileState(t, fix.Runtime, inst.ID, 4, 0, 0)
	go mirrorVirtualFiles(ctx, fix.Admin.DB, proj.ID, repoDir)

	adm, release, err := fix.Runtime.Admin(ctx, inst.ID)
	require.NoError(t, err)
	defer release()

	alertOpts := func(emails ...string) *adminv1.AlertOptions {
		return &adminv1.AlertOptions{
			DisplayName:     "Revenue drop",
			MetricsViewName: "mv1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   `{"metrics_view":"mv1","measures":[{"name":"revenue"}],"having":{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":0}]}}}`,
			EmailRecipients: emails,
			WebOpenPath:     "/explore/e1",
		}
	}

	t.Run("creates an alert owned by the user that notifies them", func(t *testing.T) {
		name, url, err := adm.CreateAlertForUser(ctx, viewerUser.ID, alertOpts())
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/alerts/%s", fix.Admin.URLs.Frontend(), org.Organization.Name, project.Project.Name, name), url)

		spec := testruntime.GetResource(t, fix.Runtime, inst.ID, runtime.ResourceKindAlert, name).GetAlert().Spec
		require.Equal(t, viewerUser.ID, spec.Annotations["admin_owner_user_id"])
		require.Equal(t, viewerUser.ID, spec.GetQueryForUserId())
		require.Len(t, spec.Notifiers, 1)
		require.Equal(t, []any{viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])
	})

	t.Run("only sends to project members", func(t *testing.T) {
		// Other project members are allowed
		_, _, err := adm.CreateAlertForUser(ctx, viewerUser.ID, alertOpts(viewerUser.Email, ownerUser.Email))
		require.NoError(t, err)

		_, _, err = adm.CreateAlertForUser(ctx, viewerUser.ID, alertOpts(outsiderUser.Email))
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, _, err = adm.CreateAlertForUser(ctx, viewerUser.ID, alertOpts("someone@example.com"))
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		opts := alertOpts()
		opts.SlackChannels = []string{"alerts"}
		_, _, err = adm.CreateAlertForUser(ctx, viewerUser.ID, opts)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("only acts for users of the project, and only from its primary deployment", func(t *testing.T) {
		_, _, err := adm.CreateAlertForUser(ctx, outsiderUser.ID, alertOpts())
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, _, err = adm.CreateAlertForUser(ctx, "ext_8c1f0c2b", alertOpts())
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = viewer.CreateAlertForUser(ctx, &adminv1.CreateAlertForUserRequest{ProjectId: proj.ID, UserId: viewerUser.ID, Options: alertOpts()})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("deletes only the user's own alerts", func(t *testing.T) {
		theirs, err := owner.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: org.Organization.Name, Project: project.Project.Name, Options: alertOpts()})
		require.NoError(t, err)
		err = adm.DeleteAlertForUser(ctx, viewerUser.ID, theirs.Name)
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		name, _, err := adm.CreateAlertForUser(ctx, viewerUser.ID, alertOpts())
		require.NoError(t, err)
		require.NoError(t, adm.DeleteAlertForUser(ctx, viewerUser.ID, name))
		_, err = ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: name}, false)
		require.ErrorIs(t, err, drivers.ErrResourceNotFound)
	})

	t.Run("creates and deletes reports", func(t *testing.T) {
		opts := &adminv1.ReportOptions{
			DisplayName:   "Weekly revenue",
			RefreshCron:   "0 9 * * 1",
			QueryName:     "MetricsViewAggregation",                                 // nolint:staticcheck // backwards compatibility
			QueryArgsJson: `{"metrics_view":"mv1","measures":[{"name":"revenue"}]}`, // nolint:staticcheck // backwards compatibility
			ExportFormat:  runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			Explore:       "e1",
			WebOpenMode:   "recipient",
		}
		name, url, err := adm.CreateReportForUser(ctx, viewerUser.ID, opts)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/reports/%s", fix.Admin.URLs.Frontend(), org.Organization.Name, project.Project.Name, name), url)
		spec := testruntime.GetResource(t, fix.Runtime, inst.ID, runtime.ResourceKindReport, name).GetReport().Spec
		require.Equal(t, viewerUser.ID, spec.Annotations["admin_owner_user_id"])
		require.Equal(t, []any{viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])

		opts.EmailRecipients = []string{outsiderUser.Email}
		_, _, err = adm.CreateReportForUser(ctx, viewerUser.ID, opts)
		require.Equal(t, codes.InvalidArgument, status.Code(err))

		require.Equal(t, codes.PermissionDenied, status.Code(adm.DeleteReportForUser(ctx, ownerUser.ID, name)))
		require.NoError(t, adm.DeleteReportForUser(ctx, viewerUser.ID, name))
	})
}

// mirrorVirtualFiles copies a project's virtual files to a runtime's repo directory until ctx is done,
// like the repo of a deployed runtime syncs them.
func mirrorVirtualFiles(ctx context.Context, db database.DB, projectID, repoDir string) {
	for {
		files, err := db.FindVirtualFiles(ctx, projectID, "prod", time.Time{}, "", 1000)
		if err != nil {
			return
		}
		for _, vf := range files {
			path := filepath.Join(repoDir, vf.Path)
			if vf.Deleted {
				_ = os.Remove(path)
				continue
			}
			if data, err := os.ReadFile(path); err == nil && string(data) == string(vf.Data) {
				continue
			}
			_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			_ = os.WriteFile(path, vf.Data, 0o644)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
}
