package server_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// TestGetIFrameOpen tests the access of an iframe that opens a report or alert created from an embed.
// The embed user writes a report's query, which the report's transitive access is derived from, so the iframe must not be limited to the report itself.
// Without navigation, it is limited the same way as when the report runs:
// to the resources of the embed the report was created from, or else by the resource restrictions of the user's identity.
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
		// src is the iframe's URL.
		src *url.URL
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
			src:         u,
			accessToken: resp.AccessToken,
			claims:      claims.Claims(inst.ID),
			adminToken:  adminToken,
			resources:   tkn.Resources,
		}, nil
	}
	openKind := func(t *testing.T, kind embedopen.Kind, name string) string {
		res, err := embedopen.Encode(&embedopen.Payload{Kind: kind, Name: name})
		require.NoError(t, err)
		return res
	}
	open := func(t *testing.T, name string) string {
		return openKind(t, embedopen.KindReport, name)
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

	// reportYAML returns a report as the admin service saves it when an embed user creates it.
	// The report is owned by the embed user (identified by the subject of its runtime JWTs), who chooses its query and the explore it opens.
	// It is limited to the explore if it was created from an iframe that is limited to the explore.
	reportYAML := func(ownerID, metricsView, explore string, limited bool) string {
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
  explore: %s
`, metricsView, ownerID, explore)
		if limited {
			res += fmt.Sprintf("  admin_embed_resources: '[{\"type\":\"%s\",\"name\":\"%s\"}]'\n", runtime.ResourceKindExplore, explore)
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
		// Created from the limited iframe; queries the metrics view of its explore.
		"reports/r1.yaml": reportYAML(ownerID, "mv1", "e1", true),
		// Created from the limited iframe; queries a metrics view outside of its explore.
		"reports/r2.yaml": reportYAML(ownerID, "mv2", "e1", true),
		// Created from the iframe with navigation, so it is not limited.
		"reports/r3.yaml": reportYAML(ownerID, "mv2", "e2", false),
	}
	testruntime.PutFiles(t, fix.Runtime, inst.ID, reports)
	testruntime.ReconcileParserAndWait(t, fix.Runtime, inst.ID)
	testruntime.RequireReconcileState(t, fix.Runtime, inst.ID, 9, 0, 0)
	for path, data := range reports {
		createVirtualFile(t, ownerID, path, data)
	}

	t.Run("limited like the report without navigation", func(t *testing.T) {
		// The reports are limited to the explore of the iframe they were created from, so opening them gives the same access as that iframe.
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

	t.Run("runtime access when opening a limited report", func(t *testing.T) {
		// The report's query targets a metrics view outside of the explore that the report is limited to.
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

	t.Run("a report that is not limited has the access of the user's identity", func(t *testing.T) {
		// The report was created from an iframe that is not limited, so there are no resources to limit the iframe to.
		// The iframe is then only restricted by the user's identity, same as when the report runs.
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r3")})
		require.NoError(t, err)
		requireSameAccess(t, navigation, got)
	})

	t.Run("only the creator can open a report", func(t *testing.T) {
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-2", Open: open(t, "r1")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r-missing")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("report name can't be a path", func(t *testing.T) {
		// The name must not resolve to another kind of virtual file of the embed user, whose content isn't written by the admin service.
		// Otherwise, the embed user could choose the resources that the iframe is limited to.
		createVirtualFile(t, ownerID, "personal/c1.yaml", reportYAML(ownerID, "mv2", "e2", true))
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "../personal/c1")})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("project member keeps their resource restrictions", func(t *testing.T) {
		// A project member with restricted resources gets iframes with navigation, so their reports are not limited,
		// and the member chooses the dashboard that a report opens.
		// Opening a report must not replace the member's restrictions with the report's dashboard.
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

		// One report on the dashboard the member can access, and one on another dashboard.
		memberID := restricted.claims.UserID
		createVirtualFile(t, memberID, "reports/r-member-e1.yaml", reportYAML(memberID, "mv1", "e1", false))
		createVirtualFile(t, memberID, "reports/r-member-e2.yaml", reportYAML(memberID, "mv2", "e2", false))
		for _, name := range []string{"r-member-e1", "r-member-e2"} {
			for _, nav := range []bool{true, false} {
				got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, name), Navigation: nav})
				require.NoError(t, err)
				requireSameAccess(t, restricted, got)
			}
		}

		// The iframe that opens the report on the other dashboard can't access that dashboard or its metrics view.
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "member", For: forMember, Open: open(t, "r-member-e2")})
		require.NoError(t, err)
		_, err = rt.GetExplore(runtimeCtx(got), &runtimev1.GetExploreRequest{InstanceId: inst.ID, Name: "e2"})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error: %v", err)
		resp, err := aggregate(got, "mv2")
		require.Equal(t, codes.PermissionDenied, status.Code(err), "got error %v and data %v", err, resp.GetData())
		resp, err = aggregate(got, "mv1")
		require.NoError(t, err)
		require.Len(t, resp.Data, 1)
	})

	t.Run("iframe params identify the opened resource", func(t *testing.T) {
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "r1")})
		require.NoError(t, err)
		require.Equal(t, "report", got.src.Query().Get("open_kind"))
		require.Equal(t, "r1", got.src.Query().Get("open_name"))
		require.False(t, got.src.Query().Has("open_report"))
	})

	// alertYAML returns an alert as the admin service saves it when an embed user creates it.
	// Alerts aren't checked against the embed's resources, but they are limited to them when created from a limited iframe.
	alertYAML := func(ownerID, metricsView string, limited bool) string {
		res := fmt.Sprintf(`
type: alert
display_name: Alert
refs:
- MetricsView/%s
query:
  name: MetricsViewAggregation
  args_json: '{"metrics_view":"%s","dimensions":[{"name":"country"}],"measures":[{"name":"total"}]}'
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_managed: "true"
  admin_embed_owner_id: %s
`, metricsView, metricsView, ownerID)
		if limited {
			res += fmt.Sprintf("  admin_embed_resources: '[{\"type\":\"%s\",\"name\":\"e1\"}]'\n", runtime.ResourceKindExplore)
		}
		return res
	}
	createVirtualFile(t, ownerID, "alerts/a1.yaml", alertYAML(ownerID, "mv1", true))
	createVirtualFile(t, ownerID, "alerts/a2.yaml", alertYAML(ownerID, "mv2", false))

	t.Run("open an alert", func(t *testing.T) {
		// Limited like the alert without navigation
		got, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: openKind(t, embedopen.KindAlert, "a1")})
		require.NoError(t, err)
		requireSameAccess(t, limited, got)
		require.Equal(t, "alert", got.src.Query().Get("open_kind"))
		require.Equal(t, "a1", got.src.Query().Get("open_name"))
		require.Equal(t, runtime.ResourceKindExplore, got.src.Query().Get("type"))

		// Not limited with navigation
		got, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: openKind(t, embedopen.KindAlert, "a1"), Navigation: true})
		require.NoError(t, err)
		requireSameAccess(t, navigation, got)

		// An alert that is not limited has the access of the user's identity
		got, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: openKind(t, embedopen.KindAlert, "a2")})
		require.NoError(t, err)
		requireSameAccess(t, navigation, got)
	})

	t.Run("only the creator can open an alert", func(t *testing.T) {
		_, err := getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-2", Open: openKind(t, embedopen.KindAlert, "a1")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: openKind(t, embedopen.KindAlert, "a-missing")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		// The kind matters: a report's name doesn't open an alert, and vice versa
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: openKind(t, embedopen.KindAlert, "r1")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		_, err = getIFrame(t, &adminv1.GetIFrameRequest{ExternalUserId: "user-1", Open: open(t, "a1")})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}
