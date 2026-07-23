package server_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestMagicAuthTokenAuthorizationBoundary(t *testing.T) {
	// This fixture keeps the complete boundary under test: Postgres stores the
	// magic token, admin signs a real JWT, and the embedded runtime enforces it.
	fix := testadmin.NewWithOptionalRuntime(t, true)
	ctx := t.Context()

	_, _ = fix.NewUser(t) // The first fixture user is a superuser; keep the project owner ordinary.
	manager, managerClient := fix.NewUser(t)
	creator, creatorClient := fix.NewUser(t)

	orgResp, err := managerClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: "magic-boundary"})
	require.NoError(t, err)
	orgName := orgResp.Organization.Name

	projectResp, err := managerClient.CreateProject(ctx, &adminv1.CreateProjectRequest{
		Org:        orgName,
		Project:    "project-a",
		ProdSlots:  1,
		SkipDeploy: true,
	})
	require.NoError(t, err)
	projectName := projectResp.Project.Name
	projectID := projectResp.Project.Id

	instanceID := fix.NewRuntimeInstance(t)
	deploymentResp, err := managerClient.CreateDeployment(ctx, &adminv1.CreateDeploymentRequest{
		Org:         orgName,
		Project:     projectName,
		Environment: "prod",
	})
	require.NoError(t, err)
	_, err = fix.Admin.DB.UpdateDeploymentUnsafe(ctx, deploymentResp.Deployment.Id, &database.UpdateDeploymentUnsafeOptions{
		RuntimeHost:       fix.RuntimeURL(),
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   fix.RuntimeURL(),
		Status:            database.DeploymentStatusRunning,
		StatusMessage:     "Running",
	})
	require.NoError(t, err)

	ctrl, err := fix.Runtime.Controller(ctx, instanceID)
	require.NoError(t, err)
	createMagicBoundaryResources(t, ctrl)

	db := openFixturePostgres(t, fix.DatabaseURL)
	// The editor can create tokens but cannot manage other creators' tokens. The
	// project owner remains the manager through their independent admin role.
	res, err := db.ExecContext(ctx, "UPDATE project_roles SET create_magic_auth_tokens=true, manage_magic_auth_tokens=false WHERE name=$1", database.ProjectRoleNameEditor)
	require.NoError(t, err)
	rows, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
	_, err = managerClient.AddProjectMemberUser(ctx, &adminv1.AddProjectMemberUserRequest{
		Org:     orgName,
		Project: projectName,
		Email:   creator.Email,
		Role:    database.ProjectRoleNameEditor,
	})
	require.NoError(t, err)

	restrictedResp, err := managerClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
		Org:         orgName,
		Project:     projectName,
		DisplayName: "restricted-admin",
		Resources: []*adminv1.ResourceName{{
			Type: runtime.ResourceKindAPI,
			Name: "resource-a",
		}},
	})
	require.NoError(t, err)
	restrictedID := magicTokenID(t, restrictedResp.Token)
	restrictedClient := fix.NewClient(t, restrictedResp.Token)

	t.Run("project and runtime resource boundary", func(t *testing.T) {
		// A token belongs to exactly one project, and its captured project-admin
		// attribute may shape data policy but cannot enlarge resource A's allowlist.
		managerProject, err := managerClient.GetProject(ctx, &adminv1.GetProjectRequest{Org: orgName, Project: projectName})
		require.NoError(t, err)
		managerClaims, err := fix.Audience.ParseAndValidate(managerProject.Jwt)
		require.NoError(t, err)
		require.False(t, managerClaims.Claims(instanceID).EnforceResourceAllowlist, "ordinary user JWTs must retain existing built-in access behavior")

		model, err := fix.Admin.DB.FindMagicAuthToken(ctx, restrictedID, false)
		require.NoError(t, err)
		require.Equal(t, projectID, model.ProjectID)
		require.Equal(t, true, model.Attributes["admin"], "the creator's attribute should remain captured in storage")

		project, err := restrictedClient.GetProject(ctx, &adminv1.GetProjectRequest{Org: orgName, Project: projectName})
		require.NoError(t, err)
		parsedClaims, err := fix.Audience.ParseAndValidate(project.Jwt)
		require.NoError(t, err)
		runtimeClaims := parsedClaims.Claims(instanceID)
		require.Equal(t, true, runtimeClaims.UserAttributes["admin"])
		require.True(t, runtimeClaims.EnforceResourceAllowlist)

		rtClient, err := runtimeclient.New(fix.RuntimeURL(), project.Jwt)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, rtClient.Close()) })

		_, err = rtClient.GetResource(ctx, &runtimev1.GetResourceRequest{
			InstanceId: instanceID,
			Name:       &runtimev1.ResourceName{Kind: runtime.ResourceKindAPI, Name: "resource-a"},
		})
		require.NoError(t, err, "the explicitly allowed resource should be readable")

		// Resource B, reports, and alerts all exist, so PermissionDenied proves a
		// policy boundary rather than a missing-resource accident.
		for _, name := range []*runtimev1.ResourceName{
			{Kind: runtime.ResourceKindAPI, Name: "resource-b"},
			{Kind: runtime.ResourceKindReport, Name: "unlisted-report"},
			{Kind: runtime.ResourceKindAlert, Name: "unlisted-alert"},
		} {
			_, err = rtClient.GetResource(ctx, &runtimev1.GetResourceRequest{InstanceId: instanceID, Name: name})
			require.Equal(t, codes.PermissionDenied, status.Code(err), "resource %s/%s must remain outside the allowlist", name.Kind, name.Name)
		}

		_, err = managerClient.CreateProject(ctx, &adminv1.CreateProjectRequest{
			Org:        orgName,
			Project:    "project-b",
			ProdSlots:  1,
			SkipDeploy: true,
		})
		require.NoError(t, err)
		_, err = restrictedClient.GetProject(ctx, &adminv1.GetProjectRequest{Org: orgName, Project: "project-b"})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "the token must not cross its stored project boundary")
	})

	t.Run("creator and manager ownership", func(t *testing.T) {
		// Creators list and revoke only their own tokens; project token managers can
		// see and revoke every creator's token without becoming the token owner.
		creatorSelf, err := creatorClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, DisplayName: "creator-self",
			Resources: []*adminv1.ResourceName{{Type: runtime.ResourceKindAPI, Name: "resource-a"}},
		})
		require.NoError(t, err)
		creatorManaged, err := creatorClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, DisplayName: "creator-managed",
			Resources: []*adminv1.ResourceName{{Type: runtime.ResourceKindAPI, Name: "resource-a"}},
		})
		require.NoError(t, err)

		creatorList, err := creatorClient.ListMagicAuthTokens(ctx, &adminv1.ListMagicAuthTokensRequest{Org: orgName, Project: projectName, PageSize: 100})
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"creator-self", "creator-managed"}, magicTokenDisplayNames(creatorList.Tokens))

		managerList, err := managerClient.ListMagicAuthTokens(ctx, &adminv1.ListMagicAuthTokensRequest{Org: orgName, Project: projectName, PageSize: 100})
		require.NoError(t, err)
		require.Contains(t, magicTokenDisplayNames(managerList.Tokens), "restricted-admin")
		require.Contains(t, magicTokenDisplayNames(managerList.Tokens), "creator-self")
		require.Contains(t, magicTokenDisplayNames(managerList.Tokens), "creator-managed")

		_, err = creatorClient.RevokeMagicAuthToken(ctx, &adminv1.RevokeMagicAuthTokenRequest{TokenId: restrictedID})
		require.Equal(t, codes.PermissionDenied, status.Code(err), "a creator must not revoke another owner's token")
		_, err = creatorClient.RevokeMagicAuthToken(ctx, &adminv1.RevokeMagicAuthTokenRequest{TokenId: magicTokenID(t, creatorSelf.Token)})
		require.NoError(t, err, "a creator should revoke their own token")
		_, err = managerClient.RevokeMagicAuthToken(ctx, &adminv1.RevokeMagicAuthTokenRequest{TokenId: magicTokenID(t, creatorManaged.Token)})
		require.NoError(t, err, "a manager should revoke another creator's token")
	})

	t.Run("invalid writes and filter size boundary", func(t *testing.T) {
		// Invalid TTLs, resource names, and cumulative filters are rejected before
		// persistence; an exactly 1024-byte filter set remains a valid boundary case.
		before := countMagicTokenRows(t, db)
		_, err := managerClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, TtlMinutes: -1, DisplayName: "negative-ttl",
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.Equal(t, before, countMagicTokenRows(t, db), "negative TTL must not leave a row")

		_, err = managerClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, DisplayName: "empty-resource",
			Resources: []*adminv1.ResourceName{{Type: runtime.ResourceKindAPI}},
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.Equal(t, before, countMagicTokenRows(t, db), "invalid resource must not leave a row")

		_, err = managerClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, DisplayName: "filter-at-limit",
			MetricsViewFilters: map[string]*runtimev1.Expression{
				"one": expressionWithJSONSize(t, 512),
				"two": expressionWithJSONSize(t, 512),
			},
		})
		require.NoError(t, err)
		require.Equal(t, before+1, countMagicTokenRows(t, db))

		_, err = managerClient.IssueMagicAuthToken(ctx, &adminv1.IssueMagicAuthTokenRequest{
			Org: orgName, Project: projectName, DisplayName: "filter-over-limit",
			MetricsViewFilters: map[string]*runtimev1.Expression{
				"one": expressionWithJSONSize(t, 512),
				"two": expressionWithJSONSize(t, 513),
			},
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.Contains(t, err.Error(), "got 1025 bytes")
		require.Equal(t, before+1, countMagicTokenRows(t, db), "oversized cumulative filters must not leave a row")
	})

	t.Run("expiry and immediate revocation", func(t *testing.T) {
		// The service-level fixture creates an already-expired row without sleeping;
		// the public handler separately proved that users cannot request negative TTLs.
		expiredTTL := -time.Nanosecond
		expired, err := fix.Admin.IssueMagicAuthToken(ctx, &admin.IssueMagicAuthTokenOptions{
			ProjectID:       projectID,
			TTL:             &expiredTTL,
			CreatedByUserID: &manager.ID,
			Attributes:      map[string]any{"admin": true, "email": manager.Email},
			Resources:       []database.ResourceName{{Type: runtime.ResourceKindAPI, Name: "resource-a"}},
			DisplayName:     "expired",
		})
		require.NoError(t, err)
		expiredClient := fix.NewClient(t, expired.Token().String())
		_, err = expiredClient.GetProject(ctx, &adminv1.GetProjectRequest{Org: orgName, Project: projectName})
		require.Equal(t, codes.Unauthenticated, status.Code(err), "expired tokens must fail before authorization")

		// The restricted token was authenticated above and therefore cached. A
		// successful revoke must invalidate that cached credential immediately.
		_, err = managerClient.RevokeMagicAuthToken(ctx, &adminv1.RevokeMagicAuthTokenRequest{TokenId: restrictedID})
		require.NoError(t, err)
		_, err = restrictedClient.GetProject(ctx, &adminv1.GetProjectRequest{Org: orgName, Project: projectName})
		require.Equal(t, codes.Unauthenticated, status.Code(err), "revoked tokens must fail on the next request")
		_, err = fix.Admin.DB.FindMagicAuthToken(ctx, restrictedID, false)
		require.True(t, errors.Is(err, database.ErrNotFound))
	})
}

func createMagicBoundaryResources(t *testing.T, ctrl *runtime.Controller) {
	t.Helper()
	resources := []struct {
		name *runtimev1.ResourceName
		res  *runtimev1.Resource
	}{
		{
			name: &runtimev1.ResourceName{Kind: runtime.ResourceKindAPI, Name: "resource-a"},
			res:  &runtimev1.Resource{Resource: &runtimev1.Resource_Api{Api: &runtimev1.API{Spec: &runtimev1.APISpec{}}}},
		},
		{
			name: &runtimev1.ResourceName{Kind: runtime.ResourceKindAPI, Name: "resource-b"},
			res:  &runtimev1.Resource{Resource: &runtimev1.Resource_Api{Api: &runtimev1.API{Spec: &runtimev1.APISpec{}}}},
		},
		{
			name: &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "unlisted-report"},
			res:  &runtimev1.Resource{Resource: &runtimev1.Resource_Report{Report: &runtimev1.Report{Spec: &runtimev1.ReportSpec{}}}},
		},
		{
			name: &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: "unlisted-alert"},
			res:  &runtimev1.Resource{Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{Spec: &runtimev1.AlertSpec{}}}},
		},
	}
	for _, r := range resources {
		require.NoError(t, ctrl.Create(t.Context(), r.name, nil, nil, nil, nil, false, r.res))
	}
}

func openFixturePostgres(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	require.NoError(t, db.PingContext(t.Context()))
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

func countMagicTokenRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM magic_auth_tokens").Scan(&count))
	return count
}

func magicTokenID(t *testing.T, value string) string {
	t.Helper()
	token, err := authtoken.FromString(value)
	require.NoError(t, err)
	return token.ID.String()
}

func magicTokenDisplayNames(tokens []*adminv1.MagicAuthToken) []string {
	names := make([]string, len(tokens))
	for i, token := range tokens {
		names[i] = token.DisplayName
	}
	return names
}

func expressionWithJSONSize(t *testing.T, size int) *runtimev1.Expression {
	t.Helper()
	base := &runtimev1.Expression{Expression: &runtimev1.Expression_Ident{Ident: ""}}
	data, err := protojson.Marshal(base)
	require.NoError(t, err)
	require.GreaterOrEqual(t, size, len(data))

	expr := &runtimev1.Expression{Expression: &runtimev1.Expression_Ident{Ident: strings.Repeat("x", size-len(data))}}
	data, err = protojson.Marshal(expr)
	require.NoError(t, err)
	require.Len(t, data, size)
	return expr
}
