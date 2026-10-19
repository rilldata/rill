package server_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/oauthclient"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConnectorOAuthToken(t *testing.T) {
	adm := testadmin.New(t)
	ctx := t.Context()
	db := adm.Admin.DB

	user, userClient := adm.NewUser(t)
	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{Name: "oauth-org"})
	require.NoError(t, err)
	proj, err := db.InsertProject(ctx, &database.InsertProjectOptions{OrganizationID: org.ID, Name: "oauth-project"})
	require.NoError(t, err)
	depl, err := db.InsertDeployment(ctx, &database.InsertDeploymentOptions{ProjectID: proj.ID, Environment: "prod", Status: database.DeploymentStatusRunning})
	require.NoError(t, err)
	deplToken, err := adm.Admin.IssueDeploymentAuthToken(ctx, depl.ID, nil)
	require.NoError(t, err)
	deplClient := adm.NewClient(t, deplToken.Token().String())

	// Store a client for one configuration of the connector, and a user's token obtained through it
	// The revocation endpoint is unreachable (private addresses are refused), so revocation on disconnect fails.
	clientJSON, err := json.Marshal(&oauthclient.Client{ServerURL: "https://mcp.example.com", ClientID: "client-1", RevocationEndpoint: "http://127.0.0.1:1/revoke"})
	require.NoError(t, err)
	clientRow, err := db.FindOrInsertConnectorOAuthClient(ctx, &database.InsertConnectorOAuthClientOptions{
		ProjectID:  proj.ID,
		Connector:  "notion",
		ConfigHash: "hash1",
		Client:     clientJSON,
	})
	require.NoError(t, err)
	tokJSON, err := json.Marshal(&oauth2.Token{AccessToken: "access-1", RefreshToken: "refresh-1", Expiry: time.Now().Add(time.Hour)})
	require.NoError(t, err)
	_, err = db.UpsertConnectorOAuthToken(ctx, &database.UpsertConnectorOAuthTokenOptions{OAuthClientID: clientRow.ID, UserID: user.ID, Token: tokJSON})
	require.NoError(t, err)

	t.Run("only deployments can get tokens", func(t *testing.T) {
		_, err := userClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: user.ID})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("config hash is required", func(t *testing.T) {
		_, err := deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", UserId: user.ID})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("token for the authorized configuration", func(t *testing.T) {
		res, err := deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: user.ID})
		require.NoError(t, err)
		require.Equal(t, "access-1", res.AccessToken)
		require.NotNil(t, res.ExpiresOn)
	})

	t.Run("no token for another configuration", func(t *testing.T) {
		_, err := deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash2", UserId: user.ID})
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("no token for another connector", func(t *testing.T) {
		_, err := deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "linear", ConfigHash: "hash1", UserId: user.ID})
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("expired token without a refresh token is removed", func(t *testing.T) {
		expiredJSON, err := json.Marshal(&oauth2.Token{AccessToken: "access-2", Expiry: time.Now().Add(-time.Minute)})
		require.NoError(t, err)
		other, _ := adm.NewUser(t)
		_, err = db.UpsertConnectorOAuthToken(ctx, &database.UpsertConnectorOAuthTokenOptions{OAuthClientID: clientRow.ID, UserID: other.ID, Token: expiredJSON})
		require.NoError(t, err)

		_, err = deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: other.ID})
		require.Equal(t, codes.NotFound, status.Code(err))
		_, err = db.FindConnectorOAuthToken(ctx, clientRow.ID, other.ID)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("delete removes the token even if revocation fails", func(t *testing.T) {
		_, err := deplClient.DeleteConnectorOAuthToken(ctx, &adminv1.DeleteConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: user.ID})
		require.NoError(t, err)
		_, err = deplClient.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: user.ID})
		require.Equal(t, codes.NotFound, status.Code(err))

		// Deleting again is a no-op
		_, err = deplClient.DeleteConnectorOAuthToken(ctx, &adminv1.DeleteConnectorOAuthTokenRequest{Connector: "notion", ConfigHash: "hash1", UserId: user.ID})
		require.NoError(t, err)
	})
}
