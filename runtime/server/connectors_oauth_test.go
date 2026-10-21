package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGetConnectorOAuthConfig(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"connectors/tracker.yaml": `
type: connector
driver: mcp
url: https://mcp.example.com/mcp
auth: oauth
oauth_client_id: client-1
oauth_client_secret: secret-1
oauth_scopes: [read]
security:
  access: "'{{ .user.domain }}' = 'rilldata.com'"
`,
			"connectors/static.yaml": `
type: connector
driver: mcp
url: https://static.example.com/mcp
`,
		},
	})
	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	mgmtCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ManageInstances}})
	get := func(ctx context.Context, connector, domain string) (*runtimev1.GetConnectorOAuthConfigResponse, error) {
		attrs, err := structpb.NewStruct(map[string]any{"domain": domain})
		require.NoError(t, err)
		return srv.GetConnectorOAuthConfig(ctx, &runtimev1.GetConnectorOAuthConfigRequest{InstanceId: instanceID, Connector: connector, UserAttributes: attrs})
	}

	t.Run("returns the config for a user with access", func(t *testing.T) {
		res, err := get(mgmtCtx, "tracker", "rilldata.com")
		require.NoError(t, err)
		require.NotEmpty(t, res.ConfigHash)
		require.Equal(t, "https://mcp.example.com/mcp", res.Config.ResourceUrl)
		require.Equal(t, "client-1", res.Config.ClientId)
		require.Equal(t, "secret-1", res.Config.ClientSecret)
		require.Equal(t, []string{"read"}, res.Config.Scopes)
	})

	t.Run("denies a user without access", func(t *testing.T) {
		_, err := get(mgmtCtx, "tracker", "example.com")
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("requires a management token", func(t *testing.T) {
		userCtx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{UserID: "u", Permissions: []runtime.Permission{runtime.UseAI, runtime.ReadObjects}})
		_, err := get(userCtx, "tracker", "rilldata.com")
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("rejects connectors without per-user OAuth", func(t *testing.T) {
		_, err := get(mgmtCtx, "static", "rilldata.com")
		require.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	t.Run("rejects unknown connectors", func(t *testing.T) {
		_, err := get(mgmtCtx, "missing", "rilldata.com")
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}
