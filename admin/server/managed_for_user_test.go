package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"path"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

func TestSaveAlertForUser(t *testing.T) {
	ctx := t.Context()
	f := newManagedForUserFixture(t)

	props, err := structpb.NewStruct(map[string]any{"sql": "SELECT country FROM mv1 HAVING revenue < 30000", "limit": 10})
	require.NoError(t, err)
	opts := &adminv1.AlertOptions{
		DisplayName:        "Revenue drop",
		MetricsViewName:    "mv1",
		Resolver:           "metrics_sql",
		ResolverProperties: props,
		EmailRecipients:    []string{f.viewerUser.Email},
		WebOpenPath:        "/explore/e1",
	}
	save := func(c *client.Client, userID, name, version, displayName string) (*adminv1.SaveAlertForUserResponse, error) {
		o := proto.Clone(opts).(*adminv1.AlertOptions)
		o.DisplayName = displayName
		return c.SaveAlertForUser(ctx, &adminv1.SaveAlertForUserRequest{ProjectId: f.projectID, UserId: userID, Name: name, Version: version, Options: o})
	}

	theirs, err := f.admin.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: opts})
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		client *client.Client
		userID string
		alert  string
		code   codes.Code
	}{
		"a user cannot save for another user":         {f.viewer, f.adminUserID, "", codes.PermissionDenied},
		"a deployment cannot save to another project": {f.otherDepl, f.viewerUser.ID, "", codes.PermissionDenied},
		"only the primary deployment can save":        {f.branchDepl, f.viewerUser.ID, "", codes.PermissionDenied},
		"the owner must be a user":                    {f.depl, "ext_8c1f0c2b", "", codes.PermissionDenied},
		"the owner must be allowed to create alerts":  {f.depl, f.outsiderUser.ID, "", codes.PermissionDenied},
		"a user cannot edit another user's alert":     {f.depl, f.viewerUser.ID, theirs.Name, codes.PermissionDenied},
		"editing a missing alert fails":               {f.depl, f.viewerUser.ID, "missing-alert", codes.NotFound},
		"names outside the generated alphabet fail":   {f.depl, f.viewerUser.ID, "../missing-alert", codes.InvalidArgument},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := save(tc.client, tc.userID, tc.alert, "v", opts.DisplayName)
			requireCode(t, tc.code, err)
		})
	}

	t.Run("creates the alert CreateAlert creates, owned by the user", func(t *testing.T) {
		res, err := save(f.depl, f.viewerUser.ID, "", "", opts.DisplayName)
		require.NoError(t, err)
		require.Regexp(t, `^revenue-drop-[0-9a-f]{8}$`, res.Name)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/alerts/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, res.Name), res.Url)

		public, err := f.viewer.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: opts})
		require.NoError(t, err)
		require.Equal(t, f.virtualFile(t, "alerts", public.Name), f.virtualFile(t, "alerts", res.Name))
	})

	t.Run("an edit is based on a version, and fails if the alert changed since", func(t *testing.T) {
		res, err := save(f.depl, f.viewerUser.ID, "", "", opts.DisplayName)
		require.NoError(t, err)

		_, err = save(f.depl, f.viewerUser.ID, res.Name, "", "Without version")
		requireCode(t, codes.InvalidArgument, err)

		edit, err := save(f.depl, f.viewerUser.ID, res.Name, res.Version, "Edited")
		require.NoError(t, err)
		require.Equal(t, "Edited", f.virtualFile(t, "alerts", res.Name)["display_name"])

		_, err = save(f.depl, f.viewerUser.ID, res.Name, res.Version, "Stale")
		requireCode(t, codes.Aborted, err)
		require.Equal(t, "Edited", f.virtualFile(t, "alerts", res.Name)["display_name"])

		// Another write lands while the edit awaits the parser, e.g. from the alert dialog.
		// The edit returns the version it wrote, so an edit based on it doesn't overwrite the other write.
		path := "alerts/" + res.Name + ".yaml"
		f.rt.onTrigger(func() {
			vf, err := f.fix.Admin.DB.FindVirtualFile(ctx, f.projectID, "prod", path)
			require.NoError(t, err)
			require.NoError(t, f.fix.Admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
				ProjectID:   f.projectID,
				Environment: "prod",
				Path:        path,
				Data:        []byte(strings.Replace(string(vf.Data), "display_name: Mine", "display_name: Concurrent", 1)),
			}))
		})
		mine, err := save(f.depl, f.viewerUser.ID, res.Name, edit.Version, "Mine")
		require.NoError(t, err)
		_, err = save(f.depl, f.viewerUser.ID, res.Name, mine.Version, "Mine again")
		requireCode(t, codes.Aborted, err)
		require.Equal(t, "Concurrent", f.virtualFile(t, "alerts", res.Name)["display_name"])
	})

	t.Run("an edit fails if the alert changes while it's being saved", func(t *testing.T) {
		res, err := save(f.depl, f.viewerUser.ID, "", "", "Mine")
		require.NoError(t, err)

		// The alert dialog writes the alert in a transaction that commits only once the edit has checked its version and waits to write.
		path := "alerts/" + res.Name + ".yaml"
		vf, err := f.fix.Admin.DB.FindVirtualFile(ctx, f.projectID, "prod", path)
		require.NoError(t, err)
		txCtx, tx, err := f.fix.Admin.DB.NewTx(ctx, false)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		require.NoError(t, f.fix.Admin.DB.UpsertVirtualFile(txCtx, &database.InsertVirtualFileOptions{
			ProjectID:   f.projectID,
			Environment: "prod",
			Path:        path,
			Data:        []byte(strings.Replace(string(vf.Data), "display_name: Mine", "display_name: Concurrent", 1)),
		}))

		errc := make(chan error, 1)
		go func() {
			_, err := save(f.depl, f.viewerUser.ID, res.Name, res.Version, "Edited")
			errc <- err
		}()
		db, err := sql.Open("pgx", f.fix.DatabaseURL)
		require.NoError(t, err)
		defer db.Close()
		require.Eventually(t, func() bool {
			var waiting int
			err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND query LIKE '%virtual_files%'`).Scan(&waiting)
			return err == nil && waiting == 1
		}, 10*time.Second, 10*time.Millisecond, "the edit's write doesn't wait for the dialog's")
		require.NoError(t, tx.Commit())

		requireCode(t, codes.Aborted, <-errc)
		require.Equal(t, "Concurrent", f.virtualFile(t, "alerts", res.Name)["display_name"])
	})
}

func TestSaveReportForUser(t *testing.T) {
	ctx := t.Context()
	f := newManagedForUserFixture(t)

	props, err := structpb.NewStruct(map[string]any{"prompt": "Summarize last week's sales", "explore": "e1"})
	require.NoError(t, err)
	aiReport := func() *adminv1.ReportOptions {
		return &adminv1.ReportOptions{
			DisplayName:        "Weekly summary",
			RefreshCron:        "0 9 * * 1",
			RefreshTimeZone:    "Europe/Madrid",
			Resolver:           "ai",
			ResolverProperties: props,
			EmailRecipients:    []string{f.viewerUser.Email},
			WebOpenMode:        "recipient",
		}
	}
	save := func(userID string, opts *adminv1.ReportOptions) (*adminv1.SaveReportForUserResponse, error) {
		return f.depl.SaveReportForUser(ctx, &adminv1.SaveReportForUserRequest{ProjectId: f.projectID, UserId: userID, Options: opts})
	}

	t.Run("the owner must be allowed to create reports", func(t *testing.T) {
		_, err := save(f.outsiderUser.ID, aiReport())
		requireCode(t, codes.PermissionDenied, err)
	})

	t.Run("creates the report CreateReport creates, owned by the user", func(t *testing.T) {
		res, err := save(f.viewerUser.ID, aiReport())
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/reports/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, res.Name), res.Url)

		public, err := f.viewer.CreateReport(ctx, &adminv1.CreateReportRequest{Org: f.org, Project: f.project, Options: aiReport()})
		require.NoError(t, err)
		require.Equal(t, f.virtualFile(t, "reports", public.Name), f.virtualFile(t, "reports", res.Name))
	})

	t.Run("rejects what needs the creator's permissions, explaining why", func(t *testing.T) {
		for name, change := range map[string]func(o *adminv1.ReportOptions){
			"a recipient outside the project": func(o *adminv1.ReportOptions) { o.EmailRecipients = append(o.EmailRecipients, "nobody@example.com") },
			"slack":                           func(o *adminv1.ReportOptions) { o.SlackChannels = []string{"#sales"} },
			"an attached export": func(o *adminv1.ReportOptions) {
				o.Resolver, o.ResolverProperties = "", nil
				o.QueryName = "MetricsViewAggregation"    // nolint:staticcheck // what the report dialog sends
				o.QueryArgsJson = `{"metricsView":"mv1"}` // nolint:staticcheck // what the report dialog sends
				o.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_CSV
				o.ExportAttach = true
				o.Explore = "e1"
			},
		} {
			t.Run(name, func(t *testing.T) {
				opts := aiReport()
				change(opts)
				_, err := save(f.viewerUser.ID, opts)
				requireCode(t, codes.InvalidArgument, err)
				require.ErrorContains(t, err, "creator")

				opts.WebOpenMode = "creator"
				_, err = save(f.viewerUser.ID, opts)
				require.NoError(t, err)
			})
		}
	})
}

func TestListAlertsForUser(t *testing.T) {
	ctx := t.Context()
	f := newManagedForUserFixture(t)

	// An alert as the alert dialog saves it, and one saved on the user's behalf.
	dialog := &adminv1.AlertOptions{
		DisplayName:          "Revenue below 30k",
		MetricsViewName:      "mv1",
		QueryName:            "MetricsViewAggregation",
		QueryArgsJson:        `{"metricsView":"mv1","measures":[{"name":"revenue"}],"dimensions":[],"timeRange":{"isoDuration":"P7D","timeZone":"Europe/Madrid"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}}`,
		RefreshTimeZone:      "Europe/Madrid",
		Renotify:             true,
		RenotifyAfterSeconds: 0,
		EmailRecipients:      []string{f.viewerUser.Email},
		SlackChannels:        []string{"#sales"},
		ExportAttach:         true,
		ExportFormat:         runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		ExportIncludeHeader:  true,
		ExportLimit:          100,
		WebOpenPath:          "/explore/e1",
		WebOpenState:         "CgJlMQ==",
	}
	public, err := f.viewer.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: dialog})
	require.NoError(t, err)
	props, err := structpb.NewStruct(map[string]any{"prompt": "Tell me if sales drop unusually", "explore": "e1"})
	require.NoError(t, err)
	ai := &adminv1.AlertOptions{DisplayName: "Unusual drop", MetricsViewName: "mv1", Resolver: "ai", ResolverProperties: props, RefreshCron: "0 8 * * *", Renotify: true, RenotifyAfterSeconds: 86400, EmailRecipients: []string{f.viewerUser.Email}}
	saved, err := f.depl.SaveAlertForUser(ctx, &adminv1.SaveAlertForUserRequest{ProjectId: f.projectID, UserId: f.viewerUser.ID, Options: ai})
	require.NoError(t, err)
	// Another user's alert.
	_, err = f.admin.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: dialog})
	require.NoError(t, err)

	_, err = f.branchDepl.ListAlertsForUser(ctx, &adminv1.ListAlertsForUserRequest{ProjectId: f.projectID, UserId: f.viewerUser.ID})
	requireCode(t, codes.PermissionDenied, err)

	// The user's alerts are listed with the options they were saved with, as they can be passed back to edit them.
	res, err := f.depl.ListAlertsForUser(ctx, &adminv1.ListAlertsForUserRequest{ProjectId: f.projectID, UserId: f.viewerUser.ID})
	require.NoError(t, err)
	got := map[string]*adminv1.AlertForUser{}
	for _, a := range res.Alerts {
		got[a.Name] = a
	}
	require.Len(t, got, 2)
	// The dialog sends a time zone without a cron when the alert runs on data refreshes, and the YAML only keeps it with a cron.
	wantDialog := proto.Clone(dialog).(*adminv1.AlertOptions)
	wantDialog.RefreshTimeZone = ""
	require.True(t, proto.Equal(wantDialog, got[public.Name].Options), "got %v", got[public.Name].Options)
	require.True(t, proto.Equal(ai, got[saved.Name].Options), "got %v", got[saved.Name].Options)
	require.Equal(t, saved.Version, got[saved.Name].Version)
	require.NotEmpty(t, got[public.Name].Version)
}

func TestListReportsForUser(t *testing.T) {
	ctx := t.Context()
	f := newManagedForUserFixture(t)

	// The three kinds of reports as the dialogs save them.
	exploreData := &adminv1.ReportOptions{
		DisplayName:         "Weekly sales",
		RefreshCron:         "0 9 * * 1",
		RefreshTimeZone:     "Europe/Madrid",
		QueryName:           "MetricsViewAggregation",                                                                                                  // nolint:staticcheck // what the report dialog sends
		QueryArgsJson:       `{"metricsView":"mv1","dimensions":[{"name":"country"}],"measures":[{"name":"revenue"}],"pivotOn":["country"],"sort":[]}`, // nolint:staticcheck // what the report dialog sends
		ExportFormat:        runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
		ExportLimit:         500,
		ExportIncludeHeader: true,
		ExportAttach:        true,
		EmailRecipients:     []string{f.viewerUser.Email},
		WebOpenState:        "CgJlMQ==",
		WebOpenMode:         "creator",
		Explore:             "e1",
	}
	canvas := &adminv1.ReportOptions{
		DisplayName:     "Overview PDF",
		RefreshCron:     "0 9 * * 1",
		ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_PDF,
		EmailRecipients: []string{f.viewerUser.Email},
		WebOpenState:    "tr=7D&f.sales=country+IN+%28%27US%27%29&f.costs=country+IN+%28%27ES%27%29&tabs=main.costs&pdf_include_filters=false&pdf_all_tabs=true",
		WebOpenMode:     "recipient",
		Canvas:          "overview",
		MetricsViewFilters: map[string]*runtimev1.Expression{
			"sales": expressionpb.In(expressionpb.Identifier("country"), []*runtimev1.Expression{expressionpb.String("US")}),
			"costs": expressionpb.In(expressionpb.Identifier("country"), []*runtimev1.Expression{expressionpb.String("ES")}),
		},
	}
	props, err := structpb.NewStruct(map[string]any{"prompt": "Summarize", "explore": "e1", "time_range": map[string]any{"expression": "7D as of latest/D"}})
	require.NoError(t, err)
	ai := &adminv1.ReportOptions{DisplayName: "Weekly insights", RefreshCron: "0 9 * * 1", Resolver: "ai", ResolverProperties: props, EmailRecipients: []string{f.viewerUser.Email}, WebOpenMode: "recipient"}

	want := map[string]*adminv1.ReportOptions{}
	for _, o := range []*adminv1.ReportOptions{exploreData, canvas, ai} {
		res, err := f.viewer.CreateReport(ctx, &adminv1.CreateReportRequest{Org: f.org, Project: f.project, Options: o})
		require.NoError(t, err)
		want[res.Name] = o
	}
	_, err = f.admin.CreateReport(ctx, &adminv1.CreateReportRequest{Org: f.org, Project: f.project, Options: ai})
	require.NoError(t, err)

	_, err = f.viewer.ListReportsForUser(ctx, &adminv1.ListReportsForUserRequest{ProjectId: f.projectID, UserId: f.viewerUser.ID})
	requireCode(t, codes.PermissionDenied, err)

	res, err := f.depl.ListReportsForUser(ctx, &adminv1.ListReportsForUserRequest{ProjectId: f.projectID, UserId: f.viewerUser.ID})
	require.NoError(t, err)
	require.Len(t, res.Reports, 3)
	for _, r := range res.Reports {
		require.Contains(t, want, r.Name)
		require.True(t, proto.Equal(want[r.Name], r.Options), "%s: got %v", r.Name, r.Options)
		require.NotEmpty(t, r.Version)
	}
}

// managedForUserFixture is a project with an admin, a viewer and an outsider, its primary deployment,
// a branch deployment of the same project, and the primary deployment of another project.
type managedForUserFixture struct {
	fix          *testadmin.Fixture
	rt           *managedResourcesRuntime
	org          string
	project      string
	projectID    string
	admin        *client.Client
	adminUserID  string
	viewer       *client.Client
	viewerUser   *database.User
	outsiderUser *database.User
	depl         *client.Client
	branchDepl   *client.Client
	otherDepl    *client.Client
}

func newManagedForUserFixture(t *testing.T) *managedForUserFixture {
	ctx := t.Context()
	fix := testadmin.New(t)
	rt := newManagedResourcesRuntime(t, fix.Admin.DB)

	_, admin := fix.NewUser(t)
	org, err := admin.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	proj, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{Org: org.Organization.Name, Project: "proj1", ProdSlots: 1, SkipDeploy: true})
	require.NoError(t, err)
	viewerUser, viewer := fix.NewUser(t)
	_, err = admin.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{Org: org.Organization.Name, Project: proj.Project.Name, Email: viewerUser.Email, Role: database.ProjectRoleNameViewer})
	require.NoError(t, err)
	outsiderUser, _ := fix.NewUser(t)
	proj2, err := admin.CreateProject(ctx, &adminv1.CreateProjectRequest{Org: org.Organization.Name, Project: "proj2", ProdSlots: 1, SkipDeploy: true})
	require.NoError(t, err)

	dbProj, err := fix.Admin.DB.FindProject(ctx, proj.Project.Id)
	require.NoError(t, err)
	return &managedForUserFixture{
		fix:          fix,
		rt:           rt,
		org:          org.Organization.Name,
		project:      proj.Project.Name,
		projectID:    proj.Project.Id,
		admin:        admin,
		adminUserID:  *dbProj.CreatedByUserID,
		viewer:       viewer,
		viewerUser:   viewerUser,
		outsiderUser: outsiderUser,
		depl:         newTestDeployment(t, fix, proj.Project.Id, rt.host, true),
		branchDepl:   newTestDeployment(t, fix, proj.Project.Id, rt.host, false),
		otherDepl:    newTestDeployment(t, fix, proj2.Project.Id, rt.host, true),
	}
}

// virtualFile returns the YAML of a managed alert or report without its nonce, which changes on every write.
func (f *managedForUserFixture) virtualFile(t *testing.T, dir, name string) map[string]any {
	vf, err := f.fix.Admin.DB.FindVirtualFile(t.Context(), f.projectID, "prod", path.Join(dir, name+".yaml"))
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(vf.Data, &doc))
	delete(doc["annotations"].(map[string]any), "admin_nonce")
	return doc
}

// managedResourcesRuntime is a runtime that presents the managed alerts and reports stored in the admin database as its resources,
// as the project parser does once triggered, so the admin service can look them up and await them.
// Like the parser, it presents the files as they were when the parser was last triggered: it reads them on every trigger.
// It doesn't compare the database's timestamps with the test's clock, since the two clocks aren't in sync.
type managedResourcesRuntime struct {
	runtimev1.UnimplementedRuntimeServiceServer
	host        string
	db          database.DB
	mu          sync.Mutex
	triggerHook func()
	parsed      map[string]map[string]*database.VirtualFile // Files by path, by instance, as of the last trigger
}

func newManagedResourcesRuntime(t *testing.T, db database.DB) *managedResourcesRuntime {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	rt := &managedResourcesRuntime{host: "http://" + lis.Addr().String(), db: db}
	srv := grpc.NewServer()
	runtimev1.RegisterRuntimeServiceServer(srv, rt)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return rt
}

var managedResourceNameRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)

func (r *managedResourcesRuntime) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	var dir string
	switch req.Name.Kind {
	case runtime.ResourceKindAlert:
		dir = "alerts"
	case runtime.ResourceKindReport:
		dir = "reports"
	}
	if dir == "" || !managedResourceNameRegexp.MatchString(req.Name.Name) {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	r.mu.Lock()
	vf := r.parsed[req.InstanceId][path.Join(dir, req.Name.Name+".yaml")]
	r.mu.Unlock()
	if vf == nil || vf.Deleted {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	var doc struct {
		Annotations map[string]any `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(vf.Data, &doc); err != nil {
		return nil, err
	}
	annotations := map[string]string{}
	for k, v := range doc.Annotations {
		annotations[k] = fmt.Sprint(v)
	}

	res := &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: req.Name, SpecVersion: vf.UpdatedOn.UnixNano()}}
	if dir == "alerts" {
		res.Resource = &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{Spec: &runtimev1.AlertSpec{Annotations: annotations}}}
	} else {
		res.Resource = &runtimev1.Resource_Report{Report: &runtimev1.Report{Spec: &runtimev1.ReportSpec{Annotations: annotations}}}
	}
	return &runtimev1.GetResourceResponse{Resource: res}, nil
}

// onTrigger sets a function to run when the parser is next triggered, e.g. to write concurrently with a save that is awaiting the parser.
func (r *managedResourcesRuntime) onTrigger(fn func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.triggerHook = fn
}

func (r *managedResourcesRuntime) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	r.mu.Lock()
	hook := r.triggerHook
	r.triggerHook = nil
	r.mu.Unlock()
	if hook != nil {
		hook()
	}

	// Parse: read the project's files as they are now.
	depl, err := r.db.FindDeploymentByInstanceID(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	files, err := r.db.FindVirtualFiles(ctx, depl.ProjectID, "prod", time.Time{}, "", 10000)
	if err != nil {
		return nil, err
	}
	parsed := make(map[string]*database.VirtualFile, len(files))
	for _, vf := range files {
		parsed[vf.Path] = vf
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.parsed == nil {
		r.parsed = map[string]map[string]*database.VirtualFile{}
	}
	r.parsed[req.InstanceId] = parsed
	return &runtimev1.CreateTriggerResponse{}, nil
}
