package server

import (
	"context"
	"fmt"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	_ "github.com/rilldata/rill/admin/database/postgres"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"strconv"
	"testing"
)

func TestAdmin_RBAC(t *testing.T) {
	//---------Setup-----------//
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/postgres", host, port.Int())

	db, err := database.Open("postgres", databaseURL)
	require.NoError(t, err)
	require.NotNil(t, db)

	require.NoError(t, db.Migrate(ctx))

	// create admin and viewer users
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:       "admin@test.io",
		DisplayName: "admin",
	})
	require.NoError(t, err)
	require.NotNil(t, adminUser)

	viewerUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:       "viewer@test.io",
		DisplayName: "viewer",
	})
	require.NoError(t, err)
	require.NotNil(t, viewerUser)

	sender, err := email.NewConsoleSender(zap.NewNop(), "rill-test@rilldata.io", "")
	require.NoError(t, err)

	service := admin.NewMock(db, nil, nil, nil, email.New(sender, ""))

	// issue admin and viewer tokens
	adminAuthToken, err := service.IssueUserAuthToken(ctx, adminUser.ID, "12345678-0000-0000-0000-000000000001", "test")
	require.NoError(t, err)
	require.NotNil(t, adminAuthToken)
	adminToken := adminAuthToken.Token().String()

	viewerAuthToken, err := service.IssueUserAuthToken(ctx, viewerUser.ID, "12345678-0000-0000-0000-000000000001", "test")
	require.NoError(t, err)
	require.NotNil(t, viewerAuthToken)
	viewerToken := viewerAuthToken.Token().String()

	// create a server instance
	server := Server{
		admin:         service,
		authenticator: auth.NewMockAuthenticator(service),
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
	adminClient := adminv1.NewAdminServiceClient(adminConn)

	viwerConn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(viewerToken)))
	require.NoError(t, err)
	defer viwerConn.Close()
	viewerClient := adminv1.NewAdminServiceClient(viwerConn)

	// make a CreateOrganization request
	adminOrg, err := adminClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: "foo",
	})
	require.NoError(t, err)
	require.Equal(t, adminOrg.Organization.Name, "foo")

	// add a viewer to the organization
	res, err := adminClient.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
		Organization: adminOrg.Organization.Name,
		Email:        viewerUser.Email,
		Role:         "viewer",
	})
	require.NoError(t, err)
	require.Equal(t, false, res.PendingSignup)

	//---------Tests-----------//

	getTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test admin get org",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test viewer get org",
			viewerClient,
			false,
			codes.OK,
		},
	}
	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{
				Name: adminOrg.Organization.Name,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, adminOrg.Organization.Name, resp.Organization.Name)
		})
	}

	membersTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test admin get members org",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test viewer get members org",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for _, tt := range membersTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.ListOrganizationMembers(ctx, &adminv1.ListOrganizationMembersRequest{
				Organization: adminOrg.Organization.Name,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, 2, len(resp.Members))
		})
	}

	listOrgTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
		numOrgs int
	}{
		{
			"test admin get orgs",
			adminClient,
			false,
			codes.OK,
			1,
		},
		{
			"test viewer get orgs",
			viewerClient,
			false,
			codes.OK,
			2,
		},
	}

	for _, tt := range listOrgTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tt.numOrgs, len(resp.Organizations))
		})
	}

	addOrgMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test add member by admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test add member by viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for i, tt := range addOrgMemberTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        strconv.Itoa(i) + "@test.io",
				Role:         "viewer",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, true, resp.PendingSignup)
		})
	}

	// test add duplicate member
	t.Run("test add duplicate member", func(t *testing.T) {
		_, err := adminClient.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
			Organization: adminOrg.Organization.Name,
			Email:        viewerUser.Email,
			Role:         "viewer",
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "already member of org")
	})

	// remove user tests
	removeOrgMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test remove member admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test remove member viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for _, tt := range removeOrgMemberTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.RemoveOrganizationMember(ctx, &adminv1.RemoveOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        viewerUser.Email,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// add the user back
			_, err = adminClient.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        viewerUser.Email,
				Role:         "viewer",
			})
			require.NoError(t, err)
		})
	}

	// remove last admin tests
	t.Run("test remove last admin", func(t *testing.T) {
		_, err := adminClient.RemoveOrganizationMember(ctx, &adminv1.RemoveOrganizationMemberRequest{
			Organization: adminOrg.Organization.Name,
			Email:        adminUser.Email,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "cannot remove the last owner")
	})
	t.Run("test leave last admin", func(t *testing.T) {
		_, err := adminClient.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{
			Organization: adminOrg.Organization.Name,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "cannot remove the last owner")
	})

	t.Run("test quota single-user orgs", func(t *testing.T) {
		i := 0
		for ; i < 4; i++ {
			orgName := "org" + strconv.Itoa(i)
			org, err := viewerClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
				Name: orgName,
			})
			if err != nil {
				require.Equal(t, codes.FailedPrecondition, status.Code(err))
				require.ErrorContains(t, err, "quota exceeded")
				break
			}
			require.NoError(t, err)
			require.Equal(t, org.Organization.Name, orgName)
		}
		require.Equal(t, 3, i)
	})

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
