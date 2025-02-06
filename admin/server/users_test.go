package server

import (
	"context"
	"net"
	"testing"

	"github.com/rilldata/rill/admin/jobs"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/admin/server/cookies"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/rilldata/rill/admin/database/postgres"
	_ "github.com/rilldata/rill/admin/provisioner/static"
)

func TestUser(t *testing.T) {
	//---------Setup-----------//
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()

	// Setup an error logger
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.ErrorLevel)
	logger, err := cfg.Build()
	require.NoError(t, err)

	sender, err := email.NewConsoleSender(logger, "rill-test@rilldata.io", "")
	require.NoError(t, err)
	emailClient := email.New(sender)

	github := &mockGithub{}

	issuer, err := runtimeauth.NewEphemeralIssuer("")
	require.NoError(t, err)

	provisionerSetJSON := "{\"static\":{\"type\":\"static\",\"spec\":{\"runtimes\":[{\"host\":\"http://localhost:9091\",\"slots\":50,\"data_dir\":\"\",\"audience_url\":\"http://localhost:8081\"}]}}}"

	service, err := admin.New(context.Background(),
		&admin.Options{
			DatabaseDriver:     "postgres",
			DatabaseDSN:        pg.DatabaseURL,
			ProvisionerSetJSON: provisionerSetJSON,
			DefaultProvisioner: "static",
			ExternalURL:        "http://localhost:9090",
			VersionNumber:      "",
		},
		logger,
		issuer,
		emailClient,
		github,
		ai.NewNoop(),
		nil,
		billing.NewNoop(),
		payment.NewNoop(),
	)
	require.NoError(t, err)

	service.Jobs = jobs.NewNoopClient()

	db := service.DB

	// create admin and viewer users to test user operations
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "admin@test.io",
		DisplayName:         "admin",
		QuotaSingleuserOrgs: 3,
		Superuser:           true,
	})
	require.NoError(t, err)
	require.NotNil(t, adminUser)

	viewerUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "viewer@test.io",
		DisplayName:         "viewer",
		QuotaSingleuserOrgs: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, viewerUser)

	viewer2User, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "viewer2@test.io",
		DisplayName:         "viewer",
		QuotaSingleuserOrgs: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, viewer2User)

	// issue admin and viewer tokens
	adminAuthToken, err := service.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, adminAuthToken)
	adminToken := adminAuthToken.Token().String()

	viewerAuthToken, err := service.IssueUserAuthToken(ctx, viewerUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, viewerAuthToken)
	viewerToken := viewerAuthToken.Token().String()

	viewer2AuthToken, err := service.IssueUserAuthToken(ctx, viewerUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, viewer2AuthToken)

	authenticator, err := auth.NewAuthenticator(logger, service, cookies.New(logger, nil), &auth.AuthenticatorOptions{
		AuthDomain: "gorillio-stage.auth0.com",
	})
	require.NoError(t, err)

	// create a server instance
	server := Server{
		admin:         service,
		opts:          &Options{},
		authenticator: authenticator,
		logger:        logger,
	}

	// create a mock bufconn listener
	lis := bufconn.Listen(1024 * 1024)
	// create a server instance listening on the mock listener
	s := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			server.authenticator.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			server.authenticator.UnaryServerInterceptor(),
		))
	adminv1.RegisterAdminServiceServer(s, &server)

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
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(adminToken)))
	require.NoError(t, err)
	defer adminConn.Close()
	adminUserClient := adminv1.NewAdminServiceClient(adminConn)

	viewerConn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(viewerToken)))
	require.NoError(t, err)
	defer viewerConn.Close()
	viewerClient := adminv1.NewAdminServiceClient(viewerConn)

	viewer2Conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(viewerToken)))
	require.NoError(t, err)
	defer viewer2Conn.Close()
	viewer2Client := adminv1.NewAdminServiceClient(viewer2Conn)

	// make a CreateOrganization request
	adminOrg, err := adminUserClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: "foo",
	})
	require.NoError(t, err)
	require.Equal(t, adminOrg.Organization.Name, "foo")

	// add a viewer to the organization
	res, err := adminUserClient.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
		Organization: adminOrg.Organization.Name,
		Email:        viewerUser.Email,
		Role:         "viewer",
	})
	require.NoError(t, err)
	require.Equal(t, false, res.PendingSignup)

	//---------Tests-----------//

	// Delete user tests
	deleteUserTests := []struct {
		name         string
		client       adminv1.AdminServiceClient
		userToDelete string
		wantErr      bool
		errCode      codes.Code
	}{
		{
			"test delete another user as a viewer",
			viewerClient,
			viewer2User.Email,
			true,
			codes.PermissionDenied,
		},
		{
			"test delete self as a viewer",
			viewer2Client,
			viewer2User.Email,
			true,
			codes.PermissionDenied,
		},
		{
			"test delete another user as an admin",
			adminUserClient,
			viewerUser.Email,
			false,
			codes.OK,
		},
		{
			"test delete self as superuser",
			adminUserClient,
			adminUser.Email,
			false,
			codes.OK,
		},
		{
			"test delete with deleted user",
			adminUserClient,
			viewer2User.Email,
			true,
			codes.NotFound,
		},
	}

	for _, tt := range deleteUserTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.client.DeleteUser(ctx, &adminv1.DeleteUserRequest{
				Email: tt.userToDelete,
			})
			t.Logf("err: %v", err)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
