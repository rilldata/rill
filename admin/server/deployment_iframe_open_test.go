package server_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/embedopen"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestGetIFrameOpen tests the access of an embed user who creates and opens reports from an embed that is limited to a single dashboard.
// The embed user writes a report's query, which the report's transitive access is derived from.
// So an iframe that opens a report must get the same access as an iframe for the report's dashboard,
// and a report that is limited to a dashboard can only query the dashboard's metrics view.
func TestGetIFrameOpen(t *testing.T) {
	fix := testadmin.NewWithOptionalRuntime(t, true)
	ctx := t.Context()

	// Create a runtime instance with two explores on separate metrics views.
	// The instance is not provisioned by the admin service (which requires deploying a project),
	// so it doesn't sync virtual files, and the test adds the reports to the instance and to the admin service separately.
	metricsView := `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- name: total
  expression: count(*)
`
	repoDir := t.TempDir()
	for path, data := range map[string]string{
		"rill.yaml":        "",
		"models/bar.sql":   `SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, 'Denmark' AS country`,
		"metrics/mv1.yaml": metricsView,
		"metrics/mv2.yaml": metricsView,
		"explores/e1.yaml": "type: explore\nmetrics_view: mv1\n",
		"explores/e2.yaml": "type: explore\nmetrics_view: mv2\n",
	} {
		abs := filepath.Join(repoDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}
	inst := &drivers.Instance{
		Environment:      "prod",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		AdminConnector:   "noop_admin",
		Connectors: []*runtimev1.Connector{
			{Type: "file", Name: "repo", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": repoDir}))},
			{Type: "duckdb", Name: "duckdb", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": ":memory:", "mode": "readwrite"}))},
			{Type: "sqlite", Name: "catalog", Config: testruntime.Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s-catalog?mode=memory&cache=shared", t.Name())}))},
		},
		Variables: map[string]string{"rill.stage_changes": "false", "rill.watch_repo": "false"},
	}
	require.NoError(t, fix.Runtime.CreateInstance(ctx, inst))
	ctrl, err := fix.Runtime.Controller(ctx, inst.ID)
	require.NoError(t, err)
	require.NoError(t, ctrl.WaitUntilIdle(ctx, false))
	testruntime.RequireReconcileState(t, fix.Runtime, inst.ID, 6, 0, 0)

	// Create a project with a deployment that points to the runtime instance.
	// The first user becomes a superuser, so the second one owns the org and project to avoid confounding effects.
	_, _ = fix.NewUser(t)
	_, c := fix.NewUser(t)
	org, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	project, err := c.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        org.Organization.Name,
		Project:    "embed-open",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	depl, err := c.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
		Org:         org.Organization.Name,
		Project:     project.Project.Name,
		Environment: "prod",
	})
	require.NoError(t, err)
	_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, depl.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
		RuntimeHost:       fix.RuntimeURL(),
		RuntimeInstanceID: inst.ID,
		RuntimeAudience:   fix.RuntimeURL(),
		Status:            database.DeploymentStatusRunning,
		StatusMessage:     "Running",
	})
	require.NoError(t, err)
	proj, err := fix.Admin.DB.FindProjectByName(ctx, org.Organization.Name, project.Project.Name)
	require.NoError(t, err)

	// iframe is what an iframe gives the browser access to.
	type iframe struct {
		// accessToken is the runtime JWT, and claims are its claims.
		accessToken string
		claims      *runtime.SecurityClaims
		// adminToken is the embed user's token for the admin service, and resources are the resources it is limited to.
		adminToken string
		resources  []database.ResourceName
	}
	getIFrame := func(t *testing.T, req *adminv1.GetIFrameRequest) (*iframe, error) {
		req.Org = org.Organization.Name
		req.Project = project.Project.Name
		resp, err := c.GetIFrame(ctx, req)
		if err != nil {
			return nil, err
		}
		claims, err := fix.Audience.ParseAndValidate(resp.AccessToken)
		require.NoError(t, err)
		u, err := url.Parse(resp.IframeSrc)
		require.NoError(t, err)
		adminToken := u.Query().Get("admin_token")
		tkn, err := fix.Admin.ValidateEmbedToken(adminToken)
		require.NoError(t, err)
		return &iframe{
			accessToken: resp.AccessToken,
			claims:      claims.Claims(inst.ID),
			adminToken:  adminToken,
			resources:   tkn.Resources,
		}, nil
	}
	open := func(t *testing.T, name string) string {
		res, err := embedopen.Encode(&embedopen.Payload{Kind: embedopen.KindReport, Name: name})
		require.NoError(t, err)
		return res
	}
	requireSameAccess := func(t *testing.T, want, got *iframe) {
		require.Equal(t, want.claims.UserID, got.claims.UserID)
		require.Equal(t, want.claims.UserAttributes, got.claims.UserAttributes)
		require.Equal(t, want.claims.Permissions, got.claims.Permissions)
		require.Len(t, got.claims.AdditionalRules, len(want.claims.AdditionalRules))
		for i, rule := range want.claims.AdditionalRules {
			require.True(t, proto.Equal(rule, got.claims.AdditionalRules[i]), "rule %d: want %v, got %v", i, rule, got.claims.AdditionalRules[i])
		}
		require.Equal(t, want.resources, got.resources)
	}

	// aggregate queries a metrics view in the runtime with an iframe's runtime JWT, as the embed's UI does.
	conn, err := grpc.NewClient(strings.TrimPrefix(fix.RuntimeURL(), "http://"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	rt := runtimev1.NewRuntimeServiceClient(conn)
	queries := runtimev1.NewQueryServiceClient(conn)
	runtimeCtx := func(f *iframe) context.Context {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+f.accessToken)
	}
	aggregate := func(f *iframe, metricsView string) (*runtimev1.MetricsViewAggregationResponse, error) {
		return queries.MetricsViewAggregation(runtimeCtx(f), &runtimev1.MetricsViewAggregationRequest{
			InstanceId:  inst.ID,
			MetricsView: metricsView,
			Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
			Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "total"}},
		})
	}

	// The iframes of an embed user who is limited to the explore e1, and of the same user with navigation.
	limited, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Resource: "e1"})
	require.NoError(t, err)
	require.Equal(t, []database.ResourceName{{Type: runtime.ResourceKindExplore, Name: "e1"}}, limited.resources)
	require.Len(t, limited.claims.AdditionalRules, 1)
	require.True(t, proto.Equal(&runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"}, limited.claims.AdditionalRules[0].GetTransitiveAccess().GetResource()))
	navigation, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Navigation: true})
	require.NoError(t, err)
	require.Empty(t, navigation.resources)
	require.Empty(t, navigation.claims.AdditionalRules)
	_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-2", Resource: "e1"})
	require.NoError(t, err)

	// Add reports as the admin service saves them when the embed user creates them from the limited iframe:
	// they are owned by the embed user (identified by the subject of its runtime JWTs), and open and are limited to the explore e1.
	// r1 queries the explore's metrics view, and r2 queries a metrics view outside of the explore.
	reportYAML := func(ownerID, metricsView, explore string) string {
		res := fmt.Sprintf(`
type: report
display_name: Report
query:
  name: MetricsViewAggregation
  args_json: '{"metrics_view":"%s","dimensions":[{"name":"country"}],"measures":[{"name":"total"}]}'
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_managed: "true"
  admin_embed_owner_id: %s
  web_open_mode: creator
`, metricsView, ownerID)
		if explore != "" {
			res += fmt.Sprintf("  explore: %s\n  admin_embed_resources: '[{\"type\":\"%s\",\"name\":\"%s\"}]'\n", explore, runtime.ResourceKindExplore, explore)
		}
		return res
	}
	createVirtualFile := func(t *testing.T, ownerID, path, data string) {
		err := fix.Admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID:    proj.ID,
			Environment:  "prod",
			Path:         path,
			Data:         []byte(data),
			EmbedOwnerID: &ownerID,
		})
		require.NoError(t, err)
	}
	ownerID := limited.claims.UserID
	reports := map[string]string{
		"reports/r1.yaml": reportYAML(ownerID, "mv1", "e1"),
		"reports/r2.yaml": reportYAML(ownerID, "mv2", "e1"),
	}
	testruntime.PutFiles(t, fix.Runtime, inst.ID, reports)
	testruntime.ReconcileParserAndWait(t, fix.Runtime, inst.ID)
	testruntime.RequireReconcileState(t, fix.Runtime, inst.ID, 8, 0, 0)
	for path, data := range reports {
		createVirtualFile(t, ownerID, path, data)
	}

	t.Run("limited to the report's dashboard without navigation", func(t *testing.T) {
		for _, name := range []string{"r1", "r2"} {
			got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, name)})
			require.NoError(t, err)
			requireSameAccess(t, limited, got)
		}
	})

	t.Run("not limited with navigation", func(t *testing.T) {
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r1"), Navigation: true})
		require.NoError(t, err)
		requireSameAccess(t, navigation, got)
	})

	t.Run("runtime access when opening a report", func(t *testing.T) {
		// The report's query targets a metrics view outside of the explore that the report opens.
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r2")})
		require.NoError(t, err)
		ctx := runtimeCtx(got)

		// The metrics view of the report's query is not accessible, since it's not the metrics view of the explore.
		resp, err := aggregate(got, "mv2")
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error %v and data %v", err, resp.GetData())
		_, err = rt.GetExplore(ctx, &runtimev1.GetExploreRequest{InstanceId: inst.ID, Name: "e2"})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)

		// The open page gets the report and its explore, and queries the explore's metrics view to restore the report's state.
		// The explore's metrics view is not limited to the fields and rows of the report's query, same as in an iframe for the explore.
		_, err = rt.GetResource(ctx, &runtimev1.GetResourceRequest{InstanceId: inst.ID, Name: &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"}})
		require.NoError(t, err)
		_, err = rt.GetExplore(ctx, &runtimev1.GetExploreRequest{InstanceId: inst.ID, Name: "e1"})
		require.NoError(t, err)
		_, err = queries.MetricsViewTimeRange(ctx, &runtimev1.MetricsViewTimeRangeRequest{InstanceId: inst.ID, MetricsViewName: "mv1"})
		require.NoError(t, err)
		resp, err = aggregate(got, "mv1")
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
	})

	t.Run("report without a dashboard can't be opened without navigation", func(t *testing.T) {
		// Without a dashboard to limit the iframe to, it must not fall back to an iframe that isn't limited.
		createVirtualFile(t, ownerID, "reports/r-no-explore.yaml", reportYAML(ownerID, "mv1", ""))
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r-no-explore")})
		require.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	t.Run("only the creator can open a report", func(t *testing.T) {
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-2", Open: open(t, "r1")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r-missing")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("report name can't be a path", func(t *testing.T) {
		// The name must not resolve to another kind of virtual file of the embed user, whose content isn't written by the admin service.
		createVirtualFile(t, ownerID, "personal/c1.yaml", reportYAML(ownerID, "mv2", "e2"))
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "../personal/c1")})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("project member with restricted resources can't open a report on another dashboard", func(t *testing.T) {
		// The embed user chooses the dashboard that a report opens when their iframes are not limited to a dashboard.
		// Limiting the iframe to it must not replace the restrictions of a project member who can only access other dashboards.
		member, _ := fix.NewUser(t)
		restrict := true
		_, err := c.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
			Org:               org.Organization.Name,
			Project:           project.Project.Name,
			Email:             member.Email,
			Role:              database.ProjectRoleNameViewer,
			RestrictResources: &restrict,
			Resources:         []*adminv1.ResourceName{{Type: runtime.ResourceKindExplore, Name: "e1"}},
		})
		require.NoError(t, err)
		forMember := &adminv1.GetIFrameRequest_UserEmail{UserEmail: member.Email}
		restricted, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Navigation: true})
		require.NoError(t, err)
		require.Empty(t, restricted.resources)
		require.Len(t, restricted.claims.AdditionalRules, 1)
		require.True(t, proto.Equal(&runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"}, restricted.claims.AdditionalRules[0].GetTransitiveAccess().GetResource()))

		createVirtualFile(t, restricted.claims.UserID, "reports/r-member-e1.yaml", reportYAML(restricted.claims.UserID, "mv1", "e1"))
		createVirtualFile(t, restricted.claims.UserID, "reports/r-member-e2.yaml", reportYAML(restricted.claims.UserID, "mv2", "e2"))

		// A report on a dashboard the member can access opens with and without navigation.
		for _, nav := range []bool{true, false} {
			got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-e1"), Navigation: nav})
			require.NoError(t, err)
			require.Len(t, got.claims.AdditionalRules, 1)
			require.True(t, proto.Equal(restricted.claims.AdditionalRules[0], got.claims.AdditionalRules[0]))
		}

		// A report on another dashboard keeps the member's restrictions with navigation, and can't be opened without it.
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-e2"), Navigation: true})
		require.NoError(t, err)
		require.Len(t, got.claims.AdditionalRules, 1)
		require.True(t, proto.Equal(restricted.claims.AdditionalRules[0], got.claims.AdditionalRules[0]))
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-e2")})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)

		// The runtime identifies resources by their lowercased name, so a report that names the member's dashboard in another case opens as well.
		// Its iframe is limited to the same dashboard, and a report that names another dashboard in another case still can't be opened.
		createVirtualFile(t, restricted.claims.UserID, "reports/r-member-case-e1.yaml", reportYAML(restricted.claims.UserID, "mv1", "E1"))
		createVirtualFile(t, restricted.claims.UserID, "reports/r-member-case-e2.yaml", reportYAML(restricted.claims.UserID, "mv2", "E2"))
		got, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-case-e1")})
		require.NoError(t, err)
		resp, err := aggregate(got, "mv1")
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
		resp, err = aggregate(got, "mv2")
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error %v and data %v", err, resp.GetData())
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-case-e2")})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
	})

	// reportOptions returns the options of a report as the embed's UI creates it.
	reportOptions := func(displayName, explore, queryArgsJSON string) *adminv1.ReportOptions {
		return &adminv1.ReportOptions{
			DisplayName:     displayName,
			RefreshCron:     "0 9 * * 1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   queryArgsJSON,
			ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			EmailRecipients: []string{"embed@example.com"},
			Explore:         explore,
			WebOpenMode:     "creator",
		}
	}
	// requireReportSaved calls a function that creates or edits a report, and waits until it has saved a report with the given display name.
	// It then cancels the call, since the call would wait for the runtime to parse the report, which the fixture's runtime instance never does.
	requireReportSaved := func(t *testing.T, displayName string, fn func(ctx context.Context) error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		errCh := make(chan error, 1)
		go func() { errCh <- fn(ctx) }()
		for deadline := time.Now().Add(30 * time.Second); ; time.Sleep(50 * time.Millisecond) {
			files, err := fix.Admin.DB.FindVirtualFilesByEmbedOwner(ctx, proj.ID, "prod", ownerID)
			require.NoError(t, err)
			if slices.ContainsFunc(files, func(f *database.VirtualFile) bool {
				return strings.HasPrefix(f.Path, "reports/") && strings.Contains(string(f.Data), "display_name: "+displayName)
			}) {
				return
			}
			select {
			case err := <-errCh:
				require.FailNow(t, "the call returned without saving the report", "error: %v", err)
			default:
			}
			require.True(t, time.Now().Before(deadline), "timed out waiting for the report to be saved")
		}
	}

	t.Run("limited reports can only query the metrics view of their explore", func(t *testing.T) {
		embed := fix.NewClient(t, limited.adminToken)
		create := func(ctx context.Context, c *client.Client, opts *adminv1.ReportOptions) error {
			_, err := c.CreateReport(ctx, &adminv1.CreateReportRequest{Org: org.Organization.Name, Project: project.Project.Name, Options: opts})
			return err
		}
		edit := func(ctx context.Context, c *client.Client, name string, opts *adminv1.ReportOptions) error {
			_, err := c.EditReport(ctx, &adminv1.EditReportRequest{Org: org.Organization.Name, Project: project.Project.Name, Name: name, Options: opts})
			return err
		}

		// Queries on a metrics view outside of the explore are rejected, in both JSON forms of the field name.
		for _, args := range []string{`{"metrics_view":"mv2"}`, `{"metricsView":"mv2"}`, `{"metricsView":"MV2"}`} {
			err := create(ctx, embed, reportOptions("Created", "e1", args))
			require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
			err = edit(ctx, embed, "r1", reportOptions("Edited", "e1", args))
			require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
		}
		// An explore outside of the embed is rejected regardless of the query.
		err := create(ctx, embed, reportOptions("Created", "e2", `{"metrics_view":"mv2"}`))
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
		// Queries that the runtime can't derive a metrics view from are rejected.
		err = create(ctx, embed, reportOptions("Created", "e1", `{"metrics_view":`))
		require.Equal(t, codes.InvalidArgument, status.Code(err), "got error: %v", err)
		opts := reportOptions("Created", "e1", `{"metrics_view":"mv1"}`)
		opts.QueryName = "Unknown"
		err = create(ctx, embed, opts)
		require.Equal(t, codes.InvalidArgument, status.Code(err), "got error: %v", err)
		// The limit of a report applies to everyone who edits it, since it is the report that can't access other metrics views.
		err = edit(ctx, c, "r1", reportOptions("Edited", "e1", `{"metrics_view":"mv2"}`))
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)

		// Queries on the explore's metrics view are accepted, with args as the UI sends them.
		args := `{"instanceId":"default","metricsView":"mv1","dimensions":[{"name":"country"}],"measures":[{"name":"total"}],"timeRange":{"isoDuration":"P7D"}}`
		requireReportSaved(t, "Created from limited embed", func(ctx context.Context) error {
			return create(ctx, embed, reportOptions("Created from limited embed", "e1", args))
		})
		requireReportSaved(t, "Edited from limited embed", func(ctx context.Context) error {
			return edit(ctx, embed, "r1", reportOptions("Edited from limited embed", "e1", args))
		})

		// The runtime identifies resources by their lowercased name, so the explore's metrics view can be named in another case.
		// The limited iframe can query it under that name, so a report that does is accepted as well.
		resp, err := aggregate(limited, "MV1")
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
		resp, err = aggregate(limited, "MV2")
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error %v and data %v", err, resp.GetData())
		args = `{"metricsView":"MV1","dimensions":[{"name":"country"}],"measures":[{"name":"total"}]}`
		requireReportSaved(t, "Created with another case", func(ctx context.Context) error {
			return create(ctx, embed, reportOptions("Created with another case", "e1", args))
		})
		requireReportSaved(t, "Edited with another case", func(ctx context.Context) error {
			return edit(ctx, embed, "r1", reportOptions("Edited with another case", "e1", args))
		})

		// An embed that names its explore in another case is limited to the same explore, so its reports are limited to the same metrics view.
		otherCase, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Resource: "E1"})
		require.NoError(t, err)
		embed = fix.NewClient(t, otherCase.adminToken)
		err = create(ctx, embed, reportOptions("Created", "E1", `{"metricsView":"mv2"}`))
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
		requireReportSaved(t, "Created from embed with another case", func(ctx context.Context) error {
			return create(ctx, embed, reportOptions("Created from embed with another case", "E1", args))
		})
	})

	t.Run("reports that are not limited can query any metrics view", func(t *testing.T) {
		embed := fix.NewClient(t, navigation.adminToken)
		requireReportSaved(t, "Created with navigation", func(ctx context.Context) error {
			_, err := embed.CreateReport(ctx, &adminv1.CreateReportRequest{
				Org:     org.Organization.Name,
				Project: project.Project.Name,
				Options: reportOptions("Created with navigation", "e1", `{"metrics_view":"mv2"}`),
			})
			return err
		})
	})
}
