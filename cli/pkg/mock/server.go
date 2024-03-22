package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/server"
	admincli "github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

func AdminService(ctx context.Context, logger *zap.Logger, databaseURL string) (*admin.Service, error) {
	sender, err := email.NewConsoleSender(logger, "rill-test@rilldata.io", "")
	if err != nil {
		return nil, err
	}

	emailClient := email.New(sender)

	gh := &mockGithub{}
	issuer, err := runtimeauth.NewEphemeralIssuer("")
	if err != nil {
		return nil, err
	}

	provisionerSetJSON := "{\"static\":{\"type\":\"static\",\"spec\":{\"runtimes\":[{\"host\":\"http://localhost:9091\",\"slots\":50,\"data_dir\":\"\",\"audience_url\":\"http://localhost:8081\"}]}}}"

	// Init admin service
	admOpts := &admin.Options{
		DatabaseDriver:     "postgres",
		DatabaseDSN:        databaseURL,
		ProvisionerSetJSON: provisionerSetJSON,
		DefaultProvisioner: "static",
		ExternalURL:        "http://localhost:9090",
		VersionNumber:      "",
	}

	adm, err := admin.New(ctx, admOpts, logger, issuer, emailClient, gh, ai.NewNoop())
	if err != nil {
		return nil, err
	}

	return adm, nil
}

func AdminServer(ctx context.Context, logger *zap.Logger, adm *admin.Service) (*server.Server, error) {
	issuer, err := runtimeauth.NewEphemeralIssuer("")
	if err != nil {
		return nil, err
	}

	// Creating a dummy config
	seesionKeyPairs := []string{"7938b8c95ac90b3731c353076daeae8a", "90c22a5a6c6b442afdb46855f95eb7d6"}
	conf := &admincli.Config{
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

	limiter := ratelimit.NewNoop()
	srv, err := server.New(logger, adm, issuer, limiter, activity.NewNoopClient(), &server.Options{
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
		return nil, err
	}
	return srv, nil
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

func CheckServerStatus(cctx context.Context) error {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(cctx, 60*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server took too long to start")
		default:
			resp, err := client.Get("http://localhost:8080/v1/ping")
			if err == nil {
				resp.Body.Close()
				return nil
			}
			time.Sleep(100 * time.Millisecond) // Wait and retry
		}
	}
}
