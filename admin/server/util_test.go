package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs/river"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// newTestUser creates a new user using a server created with newTestServer.
func newTestUser(t *testing.T, svr *Server) (*database.User, *client.Client) {
	return newTestUserWithDomain(t, svr, "test-user.com")
}

// newTestUserWithDomain creates a new user with a random email with the given email domain using a server created with newTestServer.
func newTestUserWithDomain(t *testing.T, svr *Server, domain string) (*database.User, *client.Client) {
	rand := randomBytes(16)
	email := fmt.Sprintf("test-%x@%s", rand, domain)
	return newTestUserWithEmail(t, svr, email)
}

// newTestUserWithEmail creates a new user with the given email using a server created with newTestServer.
func newTestUserWithEmail(t *testing.T, svr *Server, email string) (*database.User, *client.Client) {
	name := fmt.Sprintf("Test %s", strings.Split(email, "@")[0])

	u, err := svr.admin.CreateOrUpdateUser(context.Background(), email, name, "")
	require.NoError(t, err)

	tkn, err := svr.admin.IssueUserAuthToken(context.Background(), u.ID, database.AuthClientIDRillWeb, "Test session", nil, nil)
	require.NoError(t, err)

	return u, newTestClient(t, svr, tkn.Token().String())
}

// newTestClient creates a new client for a server created with newTestServer.
func newTestClient(t *testing.T, svr *Server, token string) *client.Client {
	host := fmt.Sprintf("http://localhost:%d", svr.opts.GRPCPort)
	c, err := client.New(host, token, "test")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, c.Close()) })
	return c
}

// newTestServer creates a new admin service and server for testing.
// The server and its resources will be cleaned up when the test ends.
//
// The server has several limitations compared to a production server:
// - Cannot provision runtimes
// - Github operation are no-ops
// - Billing operations are no-ops
// - No configured metrics project
func newTestServer(t *testing.T) *Server {
	ctx := context.Background()

	// Postgres
	pg := pgtestcontainer.New(t)
	t.Cleanup(func() { pg.Terminate(t) })

	// Logger
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.ErrorLevel)
	logger, err := cfg.Build()
	require.NoError(t, err)

	// Sender
	sender := email.NewTestSender()
	require.NoError(t, err)
	emailClient := email.New(sender)

	// Application-managed column encryption keyring
	keyring, err := database.NewRandomKeyring()
	require.NoError(t, err)
	keyringJSON, err := json.Marshal(keyring)
	require.NoError(t, err)

	// Ports and external URLs
	httpPort := findPort(t)
	grpcPort := findPort(t)
	externalURL := fmt.Sprintf("http://localhost:%d", grpcPort)
	externalHTTPURL := fmt.Sprintf("http://localhost:%d", httpPort)
	frontendURL := "http://frontend.mock"

	// JWT issuer
	issuer, err := runtimeauth.NewEphemeralIssuer(externalHTTPURL)
	require.NoError(t, err)

	// Runtime provisioner.
	// NOTE: Only gives the appearance of a static runtime, but does not actually start one.
	// TODO: Support actually starting a runtime.
	runtimeExternalURL := "http://localhost:9091"
	runtimeAudienceURL := "http://localhost:8081"
	defaultProvisioner := "static"
	provisionerSetJSON := must(json.Marshal(map[string]any{
		"static": map[string]any{
			"type": "static",
			"spec": map[string]any{
				"runtimes": []map[string]any{
					{
						"host":         runtimeExternalURL,
						"slots":        1000000,
						"audience_url": runtimeAudienceURL,
					},
				},
			},
		},
	}))

	// Admin service
	admOpts := &admin.Options{
		DatabaseDriver:            "postgres",
		DatabaseDSN:               pg.DatabaseURL,
		DatabaseEncryptionKeyring: string(keyringJSON),
		ExternalURL:               externalURL,
		FrontendURL:               frontendURL,
		ProvisionerSetJSON:        string(provisionerSetJSON),
		DefaultProvisioner:        defaultProvisioner,
		VersionNumber:             "",
		VersionCommit:             "",
		MetricsProjectOrg:         "",
		MetricsProjectName:        "",
		AutoscalerCron:            "",
		ScaleDownConstraint:       0,
	}
	adm, err := admin.New(ctx, admOpts, logger, issuer, emailClient, &mockGithub{}, ai.NewNoop(), nil, billing.NewNoop(), payment.NewNoop())
	require.NoError(t, err)
	t.Cleanup(func() { adm.Close() })

	// Background jobs
	jobs, err := river.New(ctx, pg.DatabaseURL, adm)
	require.NoError(t, err)
	t.Cleanup(func() { jobs.Close(ctx) })
	adm.Jobs = jobs

	// Server
	srvOpts := &Options{
		HTTPPort:         httpPort,
		GRPCPort:         grpcPort,
		AllowedOrigins:   []string{"*"},
		SessionKeyPairs:  [][]byte{randomBytes(16), randomBytes(16)},
		ServePrometheus:  true,
		AuthDomain:       "gorillio-stage.auth0.com",
		AuthClientID:     "",
		AuthClientSecret: "",
	}
	srv, err := New(logger, adm, issuer, ratelimit.NewNoop(), activity.NewNoopClient(), srvOpts)
	require.NoError(t, err)

	// Serve
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return srv.ServeGRPC(ctx) })
	group.Go(func() error { return srv.ServeHTTP(ctx) })
	require.NoError(t, srv.AwaitServing(ctx))

	return srv
}

// mockGithub provides a mock implementation of admin.Github.
type mockGithub struct{}

func (m *mockGithub) AppClient() *github.Client {
	return nil
}

func (m *mockGithub) InstallationClient(installationID int64) (*github.Client, error) {
	return nil, nil
}

func (m *mockGithub) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	return "", nil
}

func findPort(t *testing.T) int {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()
	return lis.Addr().(*net.TCPAddr).Port
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
