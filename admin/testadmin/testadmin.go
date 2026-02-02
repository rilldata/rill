package testadmin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
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
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeserver "github.com/rilldata/rill/runtime/server"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	// Register drivers
	_ "github.com/rilldata/rill/admin/database/postgres"
	_ "github.com/rilldata/rill/admin/provisioner/static"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/mock/ai"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
)

// Fixture is a test fixture for an admin service and server.
// It wraps an admin service with a server running on a random port backed by a testcontainer Postgres database.
// If startRt is set to true then it also includes a runtime service and server.
// The service, servers and other resources will be cleaned up when the test that created the Fixture stops.
//
// The service has several limitations compared to a production server:
// - Github operation are no-ops in short testing mode
// - Billing operations are no-ops
// - No configured metrics project
// - Does not run background jobs
type Fixture struct {
	Admin      *admin.Service
	Server     *server.Server
	ServerOpts *server.Options
	Audience   *runtimeauth.Audience

	Runtime           *runtime.Runtime
	RuntimeServer     *runtimeserver.Server
	RuntimeServerOpts *runtimeserver.Options
}

// New creates an ephemeral admin service and server for testing.
// See the docstring for the returned Fixture for details.
func New(t *testing.T) *Fixture {
	return NewWithOptionalRuntime(t, false)
}

func NewWithOptionalRuntime(t *testing.T, startRt bool) *Fixture {
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

	// Ports and external URLs - find both ports (guaranteed to be different)
	port, runtimePort := findTwoPorts(t)
	externalURL := fmt.Sprintf("http://localhost:%d", port)
	var runtimeExternalURL string
	if startRt {
		runtimeExternalURL = fmt.Sprintf("http://localhost:%d", runtimePort)
	} else {
		runtimeExternalURL = "http://example.org"
	}
	frontendURL := "http://frontend.mock"

	// JWT issuer
	issuer, err := runtimeauth.NewEphemeralIssuer(externalURL)
	require.NoError(t, err)

	// Runtime provisioner - if startRt is false, we set up a provisioner that points to a non-existent runtime server.
	runtimeAudienceURL := runtimeExternalURL
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

	// Initialize mock AI using drivers.Open pattern
	mockAIHandle, err := drivers.Open("mock_ai", "test", map[string]any{}, storage.MustNew(os.TempDir(), nil), activity.NewNoopClient(), logger)
	require.NoError(t, err)
	t.Cleanup(func() { mockAIHandle.Close() })
	mockAI, ok := mockAIHandle.AsAI("test")
	require.True(t, ok)

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
	adm, err := admin.New(ctx, admOpts, logger, issuer, emailClient, newGithub(t), mockAI, nil, billing.NewNoop(), payment.NewNoop())
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
		SessionKeyPairs:  [][]byte{randomBytes(), randomBytes()},
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

	// Create Audience
	aud, err := runtimeauth.OpenAudience(context.Background(), zap.NewNop(), externalURL, runtimeExternalURL)
	require.NoError(t, err)

	if !startRt {
		return &Fixture{
			Admin:      adm,
			Server:     srv,
			ServerOpts: srvOpts,
			Audience:   aud,
		}
	}

	// Create and start runtime server
	rtFixture := newRuntimeServer(ctx, t, group, runtimePort, externalURL, runtimeExternalURL, logger)

	return &Fixture{
		Admin:             adm,
		Server:            srv,
		ServerOpts:        srvOpts,
		Audience:          aud,
		Runtime:           rtFixture.Runtime,
		RuntimeServer:     rtFixture.RuntimeServer,
		RuntimeServerOpts: rtFixture.RuntimeServerOpts,
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
	data := randomBytes()
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

// RuntimeURL returns the URL where the embedded runtime is accessible.
func (f *Fixture) RuntimeURL() string {
	if f.RuntimeServerOpts == nil {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d", f.RuntimeServerOpts.GRPCPort)
}

func (f *Fixture) TriggerDeployment(t *testing.T, org, project string) *database.Deployment {
	proj, err := f.Admin.DB.FindProjectByName(t.Context(), org, project)
	require.NoError(t, err)
	depl, err := f.Admin.DB.FindDeploymentsForProject(t.Context(), proj.ID, "", "")
	require.NoError(t, err)
	require.Len(t, depl, 1)
	err = river.NewReconcileDeploymentWorker(f.Admin).Work(t.Context(), &riverqueue.Job[river.ReconcileDeploymentArgs]{
		Args: river.ReconcileDeploymentArgs{
			DeploymentID: depl[0].ID,
		},
	})
	require.NoError(t, err)
	depl, err = f.Admin.DB.FindDeploymentsForProject(t.Context(), proj.ID, "", "")
	require.NoError(t, err)
	require.Len(t, depl, 1)
	return depl[0]
}

// newGithub creates a new Github client. In short testing mode this is a mock client which has no-op implementations of all methods.
// Otherwise it creates a real implementation that makes real API calls to Github.
func newGithub(t *testing.T) admin.Github {
	if testing.Short() {
		return &mockGithub{}
	}

	_, currentFile, _, _ := goruntime.Caller(0)
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

// runtimeServerFixture contains the runtime server components created for testing.
type runtimeServerFixture struct {
	Runtime           *runtime.Runtime
	RuntimeServer     *runtimeserver.Server
	RuntimeServerOpts *runtimeserver.Options
}

func newRuntimeServer(ctx context.Context, t *testing.T, group *errgroup.Group, runtimePort int, externalURL, runtimeExternalURL string, logger *zap.Logger) *runtimeServerFixture {
	// Create runtime server options
	runtimeServerOpts := &runtimeserver.Options{
		HTTPPort:        runtimePort,
		GRPCPort:        runtimePort,
		AllowedOrigins:  []string{"*"},
		SessionKeyPairs: [][]byte{randomBytes(), randomBytes()},
		AuthEnable:      true,
		AuthIssuerURL:   externalURL,        // Admin server as issuer
		AuthAudienceURL: runtimeExternalURL, // Runtime's own URL
	}

	// Create runtime
	rt := testruntime.New(t, false)

	// Create runtime server
	rtSrv, err := runtimeserver.NewServer(ctx, runtimeServerOpts, rt, logger, ratelimit.NewNoop(), activity.NewNoopClient(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { rtSrv.Close() })

	// Start runtime server in the background
	group.Go(func() error {
		return rtSrv.ServeHTTP(ctx, nil, false)
	})

	// Wait for runtime server to be ready
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("%s/v1/ping", runtimeExternalURL))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 15*time.Second, 100*time.Millisecond, "runtime server failed to start")

	return &runtimeServerFixture{
		Runtime:           rt,
		RuntimeServer:     rtSrv,
		RuntimeServerOpts: runtimeServerOpts,
	}
}

// findTwoPorts finds two different available ports.
func findTwoPorts(t *testing.T) (int, int) {
	lis1, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis1.Close()

	lis2, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis2.Close()

	port1 := lis1.Addr().(*net.TCPAddr).Port
	port2 := lis2.Addr().(*net.TCPAddr).Port

	// Since both listeners are open simultaneously, the OS guarantees different ports
	require.NotEqual(t, port1, port2, "ports should be different")

	return port1, port2
}

func randomBytes() []byte {
	b := make([]byte, 16)
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
