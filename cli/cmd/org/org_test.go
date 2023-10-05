package org

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/admin/server/cookies"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"go.uber.org/zap"
)

// Test create org with name
func TestListCmd(t *testing.T) {

	//---------Setup-----------//
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()
	logger := zap.NewNop()

	sender, err := email.NewConsoleSender(logger, "rakesh.sharma@rilldata.com", "")
	require.NoError(t, err)
	emailClient := email.New(sender, "", "")

	github := &mockGithub{}

	issuer, err := runtimeauth.NewEphemeralIssuer("")
	require.NoError(t, err)

	provisionerSpec := "{\"runtimes\":[{\"host\":\"http://localhost:9091\",\"slots\":50,\"data_dir\":\"\",\"audience_url\":\"http://localhost:8081\"}]}"

	service, err := admin.New(context.Background(),
		&admin.Options{
			DatabaseDriver:  "postgres",
			DatabaseDSN:     pg.DatabaseURL,
			ProvisionerSpec: provisionerSpec,
		},
		logger,
		issuer,
		emailClient,
		github,
	)

	fmt.Println("err", err)
	require.NoError(t, err)

	authenticator, err := auth.NewAuthenticator(logger, service, cookies.New(logger, nil), &auth.AuthenticatorOptions{
		AuthDomain: "gorillio-stage.auth0.com",
	})
	require.NoError(t, err)

	// db := service.DB

	// // create admin and viewer users
	// adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
	// 	Email:               "admin@test.io",
	// 	DisplayName:         "admin",
	// 	QuotaSingleuserOrgs: 3,
	// })
	// require.NoError(t, err)

	// // issue admin and viewer tokens
	// adminAuthToken, err := service.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	// require.NoError(t, err)
	// require.NotNil(t, adminAuthToken)
	// adminToken := adminAuthToken.Token().String()

	// // create a server instance
	// server := &server.Server{
	// 	admin:         service,
	// 	authenticator: authenticator,
	// 	logger:        logger,
	// }

	srv, err := server.New(logger, service, nil, nil, &server.Options{})

	// create a mock bufconn listener
	lis := bufconn.Listen(1024 * 1024)
	// create a server instance listening on the mock listener
	s := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			authenticator.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			authenticator.UnaryServerInterceptor(),
		))
	adminv1.RegisterAdminServiceServer(s, srv)

	defer s.Stop()

	go func() {
		err := s.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// create admin and viewer clients
	adminConn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential("")))
	require.NoError(t, err)
	defer adminConn.Close()

	adminClient := adminv1.NewAdminServiceClient(adminConn)

	fmt.Println("adminClient", adminClient)
	cmd := ListCmd(&config.Config{})
	err = cmd.Execute()
	logger.Named("Console").Info("err", zap.Error(err))
	fmt.Println("err", err)

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

type bearerTokenCredential struct {
	token string
}

func newBearerTokenCredential(token string) *bearerTokenCredential {
	return &bearerTokenCredential{
		token: token,
	}
}

func (c *bearerTokenCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + c.token, // Set the bearer token in the metadata
	}, nil
}

func (c *bearerTokenCredential) RequireTransportSecurity() bool {
	return false // false for testing
}
