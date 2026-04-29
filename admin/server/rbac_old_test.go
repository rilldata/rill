package server_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestRBACOld contains some older tests for our RBAC logic.
// New tests should be added in TestRBAC, which avoids pollution of state between different subtests.
func TestRBACOld(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	adminUser, adminClient := fix.NewUserWithEmail(t, "admin@test.io")
	viewerUser, viewerClient := fix.NewUserWithEmail(t, "viewer@test.io")

	// Create test org
	n := randomName()
	adminOrg, err := adminClient.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{Name: n})
	require.NoError(t, err)
	require.Equal(t, n, adminOrg.Organization.Name)
	require.Equal(t, adminOrg.Organization.DisplayName, "")

	// add a viewer to the organization
	res, err := adminClient.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
		Org:   adminOrg.Organization.Name,
		Email: viewerUser.Email,
		Role:  "viewer",
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
				Org: adminOrg.Organization.Name,
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
			resp, err := tt.client.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
				Org: adminOrg.Organization.Name,
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
			resp, err := tt.client.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
				Org: adminOrg.Organization.Name,
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
			resp, err := tt.client.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
				Org:   adminOrg.Organization.Name,
				Email: e,
				Role:  r,
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
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].RoleName)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)

			// clean up invite
			_, err = tt.client.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
				Org:   adminOrg.Organization.Name,
				Email: e,
			})
			require.NoError(t, err)

			// check pending invite again
			invitesResp, err = tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 0, len(invitesResp.Invites))
		})
	}

	// test add duplicate member
	t.Run("test add duplicate member", func(t *testing.T) {
		_, err = adminClient.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   adminOrg.Organization.Name,
			Email: viewerUser.Email,
			Role:  "viewer",
		})

		require.Error(t, err)
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
			_, err := adminClient.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
				Org:   adminOrg.Organization.Name,
				Email: randomEmail,
				Role:  "viewer",
			})
			require.NoError(t, err)

			// remove the user using the client under test
			resp, err := tt.client.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
				Org:   adminOrg.Organization.Name,
				Email: randomEmail,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				// clean up
				_, err = adminClient.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
					Org:   adminOrg.Organization.Name,
					Email: randomEmail,
				})
				require.NoError(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}

	// The viewer should be able to remove themselves from the org
	t.Run("remove yourself from org", func(t *testing.T) {
		_, err := viewerClient.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   adminOrg.Organization.Name,
			Email: viewerUser.Email,
		})
		require.NoError(t, err)

		// Reverse the change
		_, err = adminClient.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
			Org:   adminOrg.Organization.Name,
			Email: viewerUser.Email,
			Role:  "viewer",
		})
		require.NoError(t, err)
	})

	t.Run("test remove admin same as billing email", func(t *testing.T) {
		_, err := adminClient.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   adminOrg.Organization.Name,
			Email: adminUser.Email,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "this user is the billing email for the organization")
	})

	t.Run("test leave admin same as billing email", func(t *testing.T) {
		_, err := adminClient.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{
			Org: adminOrg.Organization.Name,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "this user is the billing email for the organization")
	})

	// remove last admin tests
	t.Run("test remove last admin", func(t *testing.T) {
		testEmail := "test@example.com"
		_, err := adminClient.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
			Org:          adminOrg.Organization.Name,
			BillingEmail: &testEmail,
		})
		require.NoError(t, err)
		_, err = adminClient.RemoveOrganizationMemberUser(ctx, &adminv1.RemoveOrganizationMemberUserRequest{
			Org:   adminOrg.Organization.Name,
			Email: adminUser.Email,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "cannot remove the last admin member")
	})

	t.Run("test leave last admin", func(t *testing.T) {
		testEmail := "test@example.com"
		_, err := adminClient.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
			Org:          adminOrg.Organization.Name,
			BillingEmail: &testEmail,
		})
		_, err = adminClient.LeaveOrganization(ctx, &adminv1.LeaveOrganizationRequest{
			Org: adminOrg.Organization.Name,
		})

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.ErrorContains(t, err, "last admin")
	})

	// Create, fetch, and delete a user group
	t.Run("test delete user group", func(t *testing.T) {
		// create
		_, err := adminClient.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  adminOrg.Organization.Name,
			Name: "group1",
		})
		require.NoError(t, err)

		// fetch
		resp, err := adminClient.GetUsergroup(ctx, &adminv1.GetUsergroupRequest{
			Org:       adminOrg.Organization.Name,
			Usergroup: "group1",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Usergroup)
		require.Equal(t, "group1", resp.Usergroup.GroupName)

		// delete
		_, err = adminClient.DeleteUsergroup(ctx, &adminv1.DeleteUsergroupRequest{
			Org:       adminOrg.Organization.Name,
			Usergroup: "group1",
		})
		require.NoError(t, err)

		// fetch again
		_, err = adminClient.GetUsergroup(ctx, &adminv1.GetUsergroupRequest{
			Org:       adminOrg.Organization.Name,
			Usergroup: "group1",
		})
		require.Error(t, err)
	})

	// Create a user group, assign an org-level role and check
	t.Run("test assign user group roles", func(t *testing.T) {
		// create
		_, err := adminClient.CreateUsergroup(ctx, &adminv1.CreateUsergroupRequest{
			Org:  adminOrg.Organization.Name,
			Name: "group2",
		})
		require.NoError(t, err)

		// assign org-level role
		_, err = adminClient.AddOrganizationMemberUsergroup(ctx, &adminv1.AddOrganizationMemberUsergroupRequest{
			Org:       adminOrg.Organization.Name,
			Usergroup: "group2",
			Role:      "viewer",
		})
		require.NoError(t, err)

		// check
		resp, err := adminClient.ListOrganizationMemberUsergroups(ctx, &adminv1.ListOrganizationMemberUsergroupsRequest{
			Org: adminOrg.Organization.Name,
		})
		require.NoError(t, err)
		require.Equal(t, 4, len(resp.Members))

		var group *adminv1.MemberUsergroup
		for _, m := range resp.Members {
			if m.GroupName == "group2" {
				group = m
				break
			}
		}
		require.NotNil(t, group)
		require.Equal(t, "viewer", group.RoleName)
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
			resp, err := tt.client.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
				Org:   adminOrg.Organization.Name,
				Email: viewerUser.Email,
				Role:  "admin",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// check role
			membersResp, err := tt.client.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 2, len(membersResp.Members))
			require.Equal(t, "admin", membersResp.Members[0].RoleName)
			require.Equal(t, "admin", membersResp.Members[1].RoleName)

			// change the role back to viewer
			resp, err = tt.client.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
				Org:   adminOrg.Organization.Name,
				Email: viewerUser.Email,
				Role:  "viewer",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// check role
			membersResp, err = tt.client.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 2, len(membersResp.Members))
			for _, m := range membersResp.Members {
				if m.UserEmail == viewerUser.Email {
					require.Equal(t, "viewer", m.RoleName)
				}
			}

			// check changing role of last admin
			_, err = tt.client.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
				Org:   adminOrg.Organization.Name,
				Email: adminUser.Email,
				Role:  "viewer",
			})
			require.Error(t, err)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.ErrorContains(t, err, "last admin")

			// check changing role of invited user
			e := "1@test.io"
			r := "viewer"
			addResp, err := tt.client.AddOrganizationMemberUser(ctx, &adminv1.AddOrganizationMemberUserRequest{
				Org:   adminOrg.Organization.Name,
				Email: e,
				Role:  r,
			})

			require.NoError(t, err)
			require.NotNil(t, addResp)
			require.Equal(t, true, addResp.PendingSignup)

			// check pending invite
			invitesResp, err := tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].RoleName)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)

			r = "admin"
			// change the role of the invited user
			_, err = tt.client.SetOrganizationMemberUserRole(ctx, &adminv1.SetOrganizationMemberUserRoleRequest{
				Org:   adminOrg.Organization.Name,
				Email: e,
				Role:  r,
			})
			require.NoError(t, err)

			// check pending invite
			invitesResp, err = tt.client.ListOrganizationInvites(ctx, &adminv1.ListOrganizationInvitesRequest{
				Org: adminOrg.Organization.Name,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(invitesResp.Invites))
			require.Equal(t, e, invitesResp.Invites[0].Email)
			require.Equal(t, r, invitesResp.Invites[0].RoleName)
			require.Equal(t, adminUser.Email, invitesResp.Invites[0].InvitedBy)
		})
	}
}
