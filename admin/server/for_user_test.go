package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/client"
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
	f := newForUserFixture(t)
	ctx := t.Context()

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
		name, url, err := f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, alertOpts())
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/alerts/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, name), url)

		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindAlert, name).GetAlert().Spec
		require.Equal(t, f.viewerUser.ID, spec.Annotations["admin_owner_user_id"])
		require.Equal(t, f.viewerUser.ID, spec.GetQueryForUserId())
		require.Len(t, spec.Notifiers, 1)
		require.Equal(t, []any{f.viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])
	})

	t.Run("only sends to project members", func(t *testing.T) {
		// Other project members are allowed
		_, _, err := f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, alertOpts(f.viewerUser.Email, f.ownerUser.Email))
		require.NoError(t, err)

		_, _, err = f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, alertOpts(f.outsiderUser.Email))
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		_, _, err = f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, alertOpts("someone@example.com"))
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		opts := alertOpts()
		opts.SlackChannels = []string{"alerts"}
		_, _, err = f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, opts)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("only acts for users of the project, and only from its primary deployment", func(t *testing.T) {
		_, _, err := f.adm.CreateAlertForUser(ctx, f.outsiderUser.ID, alertOpts())
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, _, err = f.adm.CreateAlertForUser(ctx, "ext_8c1f0c2b", alertOpts())
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = f.viewer.CreateAlertForUser(ctx, &adminv1.CreateAlertForUserRequest{ProjectId: f.proj.ID, UserId: f.viewerUser.ID, Options: alertOpts()})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("deletes only the user's own alerts", func(t *testing.T) {
		theirs, err := f.owner.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: alertOpts()})
		require.NoError(t, err)
		err = f.adm.DeleteAlertForUser(ctx, f.viewerUser.ID, theirs.Name)
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		name, _, err := f.adm.CreateAlertForUser(ctx, f.viewerUser.ID, alertOpts())
		require.NoError(t, err)
		require.NoError(t, f.adm.DeleteAlertForUser(ctx, f.viewerUser.ID, name))
		ctrl, err := f.fix.Runtime.Controller(ctx, f.instance.ID)
		require.NoError(t, err)
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
		name, url, err := f.adm.CreateReportForUser(ctx, f.viewerUser.ID, opts)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/reports/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, name), url)
		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindReport, name).GetReport().Spec
		require.Equal(t, f.viewerUser.ID, spec.Annotations["admin_owner_user_id"])
		require.Equal(t, []any{f.viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])

		opts.EmailRecipients = []string{f.outsiderUser.Email}
		_, _, err = f.adm.CreateReportForUser(ctx, f.viewerUser.ID, opts)
		require.Equal(t, codes.InvalidArgument, status.Code(err))

		require.Equal(t, codes.PermissionDenied, status.Code(f.adm.DeleteReportForUser(ctx, f.ownerUser.ID, name)))
		require.NoError(t, f.adm.DeleteReportForUser(ctx, f.viewerUser.ID, name))
	})
}

// forUserFixture is a project with a runtime instance that acts as its primary deployment, and the users of the project.
type forUserFixture struct {
	fix          *testadmin.Fixture
	org, project string
	proj         *database.Project
	instance     *drivers.Instance
	// adm is the instance's admin service, which is this admin service.
	adm          drivers.AdminService
	ownerUser    *database.User
	owner        *client.Client
	viewerUser   *database.User
	viewer       *client.Client
	outsiderUser *database.User
}

// newForUserFixture creates a project owned by a user, with another user as viewer and a third user outside the project,
// and a runtime instance whose admin service is this admin service, acting as the project's primary deployment.
// The instance has the explores "e1" and "Revenue #1" (whose name needs escaping in URLs) on the metrics view "mv1", with the measure "revenue" and the dimension "country".
func newForUserFixture(t *testing.T) *forUserFixture {
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
		"rill.yaml":                "",
		"models/bar.sql":           `SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, 'Denmark' AS country, 1 AS revenue`,
		"metrics/mv1.yaml":         "version: 1\ntype: metrics_view\nmodel: bar\ntimeseries: __time\ndimensions:\n- column: country\nmeasures:\n- name: revenue\n  expression: sum(revenue)\n",
		"explores/e1.yaml":         "type: explore\nmetrics_view: mv1\n",
		"explores/Revenue #1.yaml": "type: explore\nmetrics_view: mv1\n",
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
	}
	require.NoError(t, fix.Runtime.CreateInstance(ctx, inst))
	// Creating the instance applies the deployment's config from the admin service, which restarts its controller, so this waits for the new one.
	// A wait on the old controller never ends once it's closed, so each wait is bounded and the next try gets the new one.
	require.Eventually(t, func() bool {
		ctrl, err := fix.Runtime.Controller(ctx, inst.ID)
		if err != nil {
			return false
		}
		waitCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if ctrl.WaitUntilIdle(waitCtx, false) != nil {
			return false
		}
		rs, err := ctrl.List(ctx, "", "", false)
		return err == nil && len(rs) == 5
	}, 30*time.Second, 100*time.Millisecond)
	testruntime.RequireReconcileState(t, fix.Runtime, inst.ID, 5, 0, 0)
	go mirrorVirtualFiles(ctx, fix.Admin.DB, proj.ID, fix.Runtime, inst.ID, repoDir)

	adm, release, err := fix.Runtime.Admin(ctx, inst.ID)
	require.NoError(t, err)
	t.Cleanup(release)

	return &forUserFixture{
		fix:          fix,
		org:          org.Organization.Name,
		project:      project.Project.Name,
		proj:         proj,
		instance:     inst,
		adm:          adm,
		ownerUser:    ownerUser,
		owner:        owner,
		viewerUser:   viewerUser,
		viewer:       viewer,
		outsiderUser: outsiderUser,
	}
}

// mirrorVirtualFiles copies a project's virtual files to a runtime instance's repo directory until ctx is done,
// like the repo of a deployed runtime syncs them when the admin service triggers its project parser.
// The copy can lose the race with that trigger, so it triggers the parser itself a second after a change:
// late enough for the admin service to have looked up the resource it awaits, so it sees the change.
func mirrorVirtualFiles(ctx context.Context, db database.DB, projectID string, rt *runtime.Runtime, instanceID, repoDir string) {
	for {
		files, err := db.FindVirtualFiles(ctx, projectID, "prod", time.Time{}, "", 1000)
		if err != nil {
			return
		}
		changed := false
		for _, vf := range files {
			path := filepath.Join(repoDir, vf.Path)
			data, err := os.ReadFile(path)
			exists := err == nil
			switch {
			case vf.Deleted && exists:
				_ = os.Remove(path)
				changed = true
			case !vf.Deleted && (!exists || string(data) != string(vf.Data)):
				_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)
				_ = os.WriteFile(path, vf.Data, 0o644)
				changed = true
			}
		}
		if changed {
			time.AfterFunc(time.Second, func() {
				if ctrl, err := rt.Controller(ctx, instanceID); err == nil {
					_ = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName)
				}
			})
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(20 * time.Millisecond):
		}
	}
}
