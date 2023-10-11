package org

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/server"
	admincli "github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func TestOrgCmd(t *testing.T) {
	c := qt.New(t)
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	sender, err := email.NewConsoleSender(logger, "rill-test@rilldata.io", "")
	require.NoError(t, err)
	emailClient := email.New(sender, "", "")

	gh := &mockGithub{}
	issuer, err := runtimeauth.NewEphemeralIssuer("")
	require.NoError(t, err)

	provisionerSpec := "{\"runtimes\":[{\"host\":\"http://localhost:9091\",\"slots\":50,\"data_dir\":\"\",\"audience_url\":\"http://localhost:8081\"}]}"

	// Init admin service
	admOpts := &admin.Options{
		DatabaseDriver:  "postgres",
		DatabaseDSN:     pg.DatabaseURL,
		ProvisionerSpec: provisionerSpec,
	}

	adm, err := admin.New(ctx, admOpts, logger, issuer, emailClient, gh)
	if err != nil {
		logger.Fatal("error creating service", zap.Error(err))
	}
	defer adm.Close()

	db := adm.DB
	// create admin user
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "admin@test.io",
		DisplayName:         "admin",
		QuotaSingleuserOrgs: 3,
	})
	require.NoError(t, err)

	// issue admin and viewer tokens
	adminAuthToken, err := adm.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, adminAuthToken)

	// Creating a dummy config
	seesionKeyPairs := []string{"7938b8c95ac90b3731c353076daeae8a", "90c22a5a6c6b442afdb46855f95eb7d6"}
	conf := &admincli.Config{
		DatabaseURL:     pg.DatabaseURL,
		HTTPPort:        8080,
		GRPCPort:        9090,
		ExternalURL:     "http://localhost:8080",
		FrontendURL:     "http://localhost:3000",
		SessionKeyPairs: seesionKeyPairs,
		AuthDomain:      "gorillio-stage.auth0.com",
	}

	keyPairs := make([][]byte, len(conf.SessionKeyPairs))
	for idx, keyHex := range conf.SessionKeyPairs {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			logger.Fatal("failed to parse session key from hex string to bytes")
		}
		keyPairs[idx] = key
	}

	// Make errgroup for running the processes
	ctx = graceful.WithCancelOnTerminate(ctx)
	group, cctx := errgroup.WithContext(ctx)

	limiter := ratelimit.NewNoop()
	srv, err := server.New(logger, adm, issuer, limiter, &server.Options{
		HTTPPort:               conf.HTTPPort,
		GRPCPort:               conf.GRPCPort,
		ExternalURL:            conf.ExternalURL,
		FrontendURL:            conf.FrontendURL,
		SessionKeyPairs:        keyPairs,
		AllowedOrigins:         conf.AllowedOrigins,
		ServePrometheus:        conf.MetricsExporter == observability.PrometheusExporter,
		AuthDomain:             conf.AuthDomain,
		AuthClientID:           conf.AuthClientID,
		AuthClientSecret:       conf.AuthClientSecret,
		GithubAppName:          conf.GithubAppName,
		GithubAppWebhookSecret: conf.GithubAppWebhookSecret,
		GithubClientID:         conf.GithubClientID,
		GithubClientSecret:     conf.GithubClientSecret,
	})
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	group.Go(func() error { return srv.ServeGRPC(cctx) })
	group.Go(func() error { return srv.ServeHTTP(cctx) })

	time.Sleep(05 * time.Second)

	// Create organization with name
	var buf bytes.Buffer
	format := printer.JSON
	p := printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)

	helper := &cmdutil.Helper{
		Config: &config.Config{
			AdminURL:          "http://localhost:9090",
			AdminTokenDefault: adminAuthToken.Token().String(),
		},
		Printer: p,
	}

	cmd := CreateCmd(helper)
	cmd.UsageString()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", "myorg"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(buf.String()), &data)
	c.Assert(err, qt.IsNil)
	c.Assert(data["name"], qt.Equals, "myorg")

	// Create new organization with name
	buf.Reset()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", "test"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	err = json.Unmarshal([]byte(buf.String()), &data)
	c.Assert(err, qt.IsNil)
	c.Assert(data["name"], qt.Equals, "test")

	// List organizations
	buf.Reset()
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	orgList := []Org{}
	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 2)
	c.Assert(orgList[0].Name, qt.Equals, "myorg")
	c.Assert(orgList[1].Name, qt.Equals, "test")

	// Delete organization
	buf.Reset()
	cmd = DeleteCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", "myorg", "--force"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	// List organizations
	buf.Reset()
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	orgList = []Org{}
	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	c.Assert(orgList[0].Name, qt.Equals, "test")

	// rename organization
	buf.Reset()
	cmd = RenameCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", "test", "--new-name", "new-test", "--force"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	data = map[string]interface{}{}
	err = json.Unmarshal([]byte(buf.String()), &data)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	c.Assert(data["name"], qt.Equals, "new-test")
}

type Org struct {
	Name string `json:"Name"`
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
