package server

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/admin/server/cookies"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

func TestAdmin_RBAC(t *testing.T) {
	//---------Setup-----------//
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()
	logger := zap.NewNop()

	sender, err := email.NewConsoleSender(logger, "rill-test@rilldata.io", "")
	require.NoError(t, err)
	emailClient := email.New(sender, "")

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
	require.NoError(t, err)

	db := service.DB

	// create admin and viewer users
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "admin@test.io",
		DisplayName:         "admin",
		QuotaSingleuserOrgs: 3,
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

	testUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "test@test.io",
		DisplayName:         "test",
		QuotaSingleuserOrgs: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, testUser)

	// issue admin and viewer tokens
	adminAuthToken, err := service.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, adminAuthToken)
	adminToken := adminAuthToken.Token().String()

	viewerAuthToken, err := service.IssueUserAuthToken(ctx, viewerUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, viewerAuthToken)
	viewerToken := viewerAuthToken.Token().String()

	testAuthToken, err := service.IssueUserAuthToken(ctx, testUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, testAuthToken)
	testToken := testAuthToken.Token().String()

	authenticator, err := auth.NewAuthenticator(logger, service, cookies.New(logger, nil), &auth.AuthenticatorOptions{
		AuthDomain: "gorillio-stage.auth0.com",
	})
	require.NoError(t, err)

	// create a server instance
	server := Server{
		admin:         service,
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
	adminClient := adminv1.NewAdminServiceClient(adminConn)

	viwerConn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(viewerToken)))
	require.NoError(t, err)
	defer viwerConn.Close()
	viewerClient := adminv1.NewAdminServiceClient(viwerConn)

	testConn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(newBearerTokenCredential(testToken)))
	require.NoError(t, err)
	defer testConn.Close()
	testClient := adminv1.NewAdminServiceClient(testConn)

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
			"test get org - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test get org - viewer",
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
			"test get org members - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test get org members - viewer",
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
			"test list orgs - admin",
			adminClient,
			false,
			codes.OK,
			1,
		},
		{
			"test list orgs - viewer",
			viewerClient,
			false,
			codes.OK,
			1,
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

	// list user tests
	listOrgMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test list member - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test list member - viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for _, tt := range listOrgMemberTests {
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

		})
	}

	addOrgMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test add org member - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test add org member - viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for i, tt := range addOrgMemberTests {
		t.Run(tt.name, func(t *testing.T) {
			e := strconv.Itoa(i) + "@test.io"
			r := "viewer"
			resp, err := tt.client.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        e,
				Role:         r,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, true, resp.PendingSignup)

			// check pending invite
			invitesResp, err := tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].Role)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)

			// clean up invite
			_, err = tt.client.RemoveOrganizationMember(ctx, &adminv1.RemoveOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        e,
			})
			require.NoError(t, err)

			// check pending invite again
			invitesResp, err = tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 0, len(invitesResp.Invites))
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
		require.ErrorContains(t, err, "user is already a member of the org")
	})

	// remove user tests
	removeOrgMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"test remove member - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"test remove member - viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for _, tt := range removeOrgMemberTests {
		t.Run(tt.name, func(t *testing.T) {
			// add random user using admin client
			randomEmail := "random@test.io"
			_, err := adminClient.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        randomEmail,
				Role:         "viewer",
			})
			require.NoError(t, err)

			// remove the user using the client under test
			resp, err := tt.client.RemoveOrganizationMember(ctx, &adminv1.RemoveOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        randomEmail,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				// clean up
				_, err = adminClient.RemoveOrganizationMember(ctx, &adminv1.RemoveOrganizationMemberRequest{
					Organization: adminOrg.Organization.Name,
					Email:        randomEmail,
				})
				require.NoError(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
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

	// test change roles
	setRoleMemberTests := []struct {
		name    string
		client  adminv1.AdminServiceClient
		wantErr bool
		errCode codes.Code
	}{
		{
			"set member role - admin",
			adminClient,
			false,
			codes.OK,
		},
		{
			"set member role - viewer",
			viewerClient,
			true,
			codes.PermissionDenied,
		},
	}

	for _, tt := range setRoleMemberTests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.client.SetOrganizationMemberRole(ctx, &adminv1.SetOrganizationMemberRoleRequest{
				Organization: adminOrg.Organization.Name,
				Email:        viewerUser.Email,
				Role:         "admin",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// check role
			membersResp, err := tt.client.ListOrganizationMembers(ctx, &adminv1.ListOrganizationMembersRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 2, len(membersResp.Members))
			require.Equal(t, "admin", membersResp.Members[0].RoleName)
			require.Equal(t, "admin", membersResp.Members[1].RoleName)

			// change the role back to viewer
			resp, err = tt.client.SetOrganizationMemberRole(ctx, &adminv1.SetOrganizationMemberRoleRequest{
				Organization: adminOrg.Organization.Name,
				Email:        viewerUser.Email,
				Role:         "viewer",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// check role
			membersResp, err = tt.client.ListOrganizationMembers(ctx, &adminv1.ListOrganizationMembersRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 2, len(membersResp.Members))
			for _, m := range membersResp.Members {
				if m.UserEmail == viewerUser.Email {
					require.Equal(t, "viewer", m.RoleName)
				}
			}

			// check changing role of last admin
			_, err = tt.client.SetOrganizationMemberRole(ctx, &adminv1.SetOrganizationMemberRoleRequest{
				Organization: adminOrg.Organization.Name,
				Email:        adminUser.Email,
				Role:         "viewer",
			})
			require.Error(t, err)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.ErrorContains(t, err, "cannot change role of the last owner")

			// check changing role of invited user
			e := "1@test.io"
			r := "viewer"
			addResp, err := tt.client.AddOrganizationMember(ctx, &adminv1.AddOrganizationMemberRequest{
				Organization: adminOrg.Organization.Name,
				Email:        e,
				Role:         r,
			})

			require.NoError(t, err)
			require.NotNil(t, addResp)
			require.Equal(t, true, addResp.PendingSignup)

			// check pending invite
			invitesResp, err := tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].Role)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)

			r = "admin"
			// change the role of the invited user
			_, err = tt.client.SetOrganizationMemberRole(ctx, &adminv1.SetOrganizationMemberRoleRequest{
				Organization: adminOrg.Organization.Name,
				Email:        e,
				Role:         r,
			})
			require.NoError(t, err)

			// check pending invite
			invitesResp, err = tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Organization: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].Role)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)
		})
	}

	t.Run("test quota single-user orgs", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			orgName := "org" + strconv.Itoa(i)
			org, err := testClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
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
		resp, err := testClient.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Organizations))
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
