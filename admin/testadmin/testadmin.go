package testadmin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs/river"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ai"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	// Register database driver and supported provisioners
	_ "github.com/rilldata/rill/admin/database/postgres"
	_ "github.com/rilldata/rill/admin/provisioner/static"
)

// Fixture is a test fixture for an admin service and server.
// It wraps an admin service with a server running on a random port backed by a testcontainer Postgres database.
// The service, server and other resources will be cleaned up when the test that created the Fixture stops.
//
// The service has several limitations compared to a production server:
// - Cannot provision runtimes
// - Github operation are no-ops
// - Billing operations are no-ops
// - No configured metrics project
// - Does not run background jobs
type Fixture struct {
	Admin      *admin.Service
	Server     *server.Server
	ServerOpts *server.Options
}

// New creates an ephemeral admin service and server for testing.
// See the docstring for the returned Fixture for details.
func New(t *testing.T) *Fixture {
	ctx := t.Context()

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
	emailClient := email.New(sender)

	// Application-managed column encryption keyring
	keyring, err := database.NewRandomKeyring()
	require.NoError(t, err)
	keyringJSON, err := json.Marshal(keyring)
	require.NoError(t, err)

	// Ports and external URLs
	port := findPort(t)
	externalURL := fmt.Sprintf("http://localhost:%d", port)
	frontendURL := "http://frontend.mock"

	// JWT issuer
	issuer, err := runtimeauth.NewEphemeralIssuer(externalURL)
	require.NoError(t, err)

	// Runtime provisioner.
	// NOTE: Only gives the appearance of a static runtime, but does not actually start one.
	// TODO: Support actually starting a runtime.
	runtimeExternalURL := "http://localhost:8081"
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
		Version:                   version.Version{},
		MetricsProjectOrg:         "",
		MetricsProjectName:        "",
		AutoscalerCron:            "",
		ScaleDownConstraint:       0,
	}
	adm, err := admin.New(ctx, admOpts, logger, issuer, emailClient, newGithub(t), ai.NewNoop(), nil, billing.NewNoop(), payment.NewNoop())
	require.NoError(t, err)
	t.Cleanup(func() { adm.Close() })

	// Background jobs
	jobs, err := river.New(ctx, pg.DatabaseURL, adm)
	require.NoError(t, err)
	t.Cleanup(func() { jobs.Close(ctx) })
	adm.Jobs = jobs

	// Server
	srvOpts := &server.Options{
		HTTPPort:         port,
		GRPCPort:         port,
		AllowedOrigins:   []string{"*"},
		SessionKeyPairs:  [][]byte{randomBytes(16), randomBytes(16)},
		ServePrometheus:  true,
		AuthDomain:       "gorillio-stage.auth0.com",
		AuthClientID:     "",
		AuthClientSecret: "",
	}
	srv, err := server.New(logger, adm, issuer, ratelimit.NewNoop(), activity.NewNoopClient(), srvOpts)
	require.NoError(t, err)

	// Serve
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return srv.ServeHTTP(ctx) })
	require.NoError(t, srv.AwaitServing(ctx))

	return &Fixture{
		Admin:      adm,
		Server:     srv,
		ServerOpts: srvOpts,
	}
}

// NewUser creates a new user in the fixture's admin service.
func (f *Fixture) NewUser(t *testing.T) (*database.User, *client.Client) {
	return f.NewUserWithDomain(t, "test-user.com")
}

// NewSuperuser creates a new user with superuser permission in the fixture's admin service.
func (f *Fixture) NewSuperuser(t *testing.T) (*database.User, *client.Client) {
	u, c := f.NewUserWithDomain(t, "test-superuser.com")
	err := f.Admin.DB.UpdateSuperuser(t.Context(), u.ID, true)
	require.NoError(t, err)
	return u, c
}

// NewUserWithDomain creates a new user with a random email with the given email domain in the fixture's admin service.
func (f *Fixture) NewUserWithDomain(t *testing.T, domain string) (*database.User, *client.Client) {
	data := randomBytes(16)
	emailAddr := fmt.Sprintf("test-%x@%s", data, domain)
	return f.NewUserWithEmail(t, emailAddr)
}

// NewUserWithEmail creates a new user with the given email in the fixture's admin service.
func (f *Fixture) NewUserWithEmail(t *testing.T, emailAddr string) (*database.User, *client.Client) {
	ctx := t.Context()
	name := fmt.Sprintf("Test %s", strings.Split(emailAddr, "@")[0])

	u, err := f.Admin.CreateOrUpdateUser(ctx, emailAddr, name, "")
	require.NoError(t, err)

	tkn, err := f.Admin.IssueUserAuthToken(ctx, u.ID, database.AuthClientIDRillWeb, "Test session", nil, nil, false)
	require.NoError(t, err)

	return u, f.NewClient(t, tkn.Token().String())
}

// NewClient creates a new client for the fixture's server.
func (f *Fixture) NewClient(t *testing.T, token string) *client.Client {
	c, err := client.New(f.ExternalURL(), token, "test")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, c.Close()) })
	return c
}

// ExternalURL returns the localhost URL of the fixture's server.
func (f *Fixture) ExternalURL() string {
	return fmt.Sprintf("http://localhost:%d", f.ServerOpts.GRPCPort)
}

// newGithub creates a new Github client. In short testing mode this is a mock client which has no-op implementations of all methods.
// Otherwise it creates a real implementation that makes real API calls to Github.
func newGithub(t *testing.T) admin.Github {
	if testing.Short() {
		return &mockGithub{}
	}

	_, currentFile, _, _ := runtime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		err := godotenv.Load(envPath)
		require.NoError(t, err)
	}

	githubAppID, err := strconv.ParseInt(os.Getenv("RILL_ADMIN_TEST_GITHUB_APP_ID"), 10, 64)
	require.NoError(t, err)

	github, err := admin.NewGithub(t.Context(), githubAppID, os.Getenv("RILL_ADMIN_TEST_GITHUB_APP_PRIVATE_KEY"), os.Getenv("RILL_ADMIN_TEST_GITHUB_MANAGED_ACCOUNT"), zap.Must(zap.NewDevelopment()))
	require.NoError(t, err)
	return github
}

// mockGithub provides a mock implementation of admin.Github.
type mockGithub struct{}

func (m *mockGithub) AppClient() *github.Client {
	return nil
}

func (m *mockGithub) InstallationClient(installationID int64, repoID *int64) *github.Client {
	return nil
}

func (m *mockGithub) InstallationToken(ctx context.Context, installationID, repoID int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (m *mockGithub) InstallationTokenForOrg(ctx context.Context, org string) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (m *mockGithub) CreateManagedRepo(ctx context.Context, repoPrefix string) (*github.Repository, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockGithub) ManagedOrgInstallationID() (int64, error) {
	return 0, fmt.Errorf("not implemented")
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
