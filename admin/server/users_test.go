package server_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUser(t *testing.T) {
	ctx := context.Background()
	fix := testadmin.New(t)

	t.Run("Deleting a user", func(t *testing.T) {
		// Create a superuser and two normal users
		_, sc1 := fix.NewSuperuser(t)
		u2, c2 := fix.NewUser(t)
		u3, c3 := fix.NewUser(t)

		// A normal user can't delete another normal user
		_, err := c2.DeleteUser(ctx, &adminv1.DeleteUserRequest{
			Email: u3.Email,
		})
		require.Error(t, err)
		require.Equal(t, codes.PermissionDenied, grpc.Code(err))

		// A normal user can delete themselves
		_, err = c2.DeleteUser(ctx, &adminv1.DeleteUserRequest{
			Email: u2.Email,
		})
		require.NoError(t, err)
		fix.Admin.PurgeAuthTokenCache()

		_, err = c2.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, grpc.Code(err))

		// A superuser can delete any user
		_, err = sc1.DeleteUser(ctx, &adminv1.DeleteUserRequest{
			Email:                u3.Email,
			SuperuserForceAccess: true,
		})
		require.NoError(t, err)
		fix.Admin.PurgeAuthTokenCache()

		_, err = c3.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, grpc.Code(err))
	})

	t.Run("Single-user orgs quota", func(t *testing.T) {
		u1, c1 := fix.NewUser(t)

		_, err := fix.Admin.DB.UpdateUser(ctx, u1.ID, &database.UpdateUserOptions{
			QuotaSingleuserOrgs: 3,
		})
		require.NoError(t, err)

		for i := 0; i < 4; i++ {
			orgName := "org" + strconv.Itoa(i)
			org, err := c1.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
				Name: orgName,
			})
			if err != nil {
				require.Equal(t, codes.FailedPrecondition, status.Code(err), "error is: %v", err)
				require.ErrorContains(t, err, "quota exceeded")
				break
			}
			require.NoError(t, err)
			require.Equal(t, org.Organization.Name, orgName)
		}
		resp, err := c1.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Organizations))
	})

	t.Run("Token basics", func(t *testing.T) {
		u1, c1 := fix.NewUser(t)

		// Issue a plain token
		res, err := c1.IssueUserAuthToken(ctx, &adminv1.IssueUserAuthTokenRequest{
			UserId:      "current",
			ClientId:    database.AuthClientIDRillManual,
			DisplayName: "Foo",
		})
		require.NoError(t, err)
		require.NotEmpty(t, res.Token)

		// Check the token works
		uTmp := fix.NewClient(t, res.Token)
		res2, err := uTmp.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		require.NoError(t, err)
		require.Equal(t, res2.User.Email, u1.Email)

		// Issue a token with an expiration
		res3, err := c1.IssueUserAuthToken(ctx, &adminv1.IssueUserAuthTokenRequest{
			UserId:     "current",
			ClientId:   database.AuthClientIDRillManual,
			TtlMinutes: 10,
		})
		require.NoError(t, err)
		require.NotEmpty(t, res3.Token)

		// Check the token were created
		res4, err := c1.ListUserAuthTokens(ctx, &adminv1.ListUserAuthTokensRequest{
			UserId: "current",
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(res4.Tokens), 3) // 2 created above and 1 from fix.NewUser

		// One should have description "Foo" and one should have an expiration
		var foundFoo, foundExpiration bool
		for _, token := range res4.Tokens {
			if token.DisplayName == "Foo" {
				foundFoo = true
			}
			if token.ExpiresOn != nil {
				foundExpiration = true
			}
		}
		require.True(t, foundFoo)
		require.True(t, foundExpiration)

		// Find an ID for the "Foo" token
		var tokenID string
		for _, token := range res4.Tokens {
			if token.DisplayName == "Foo" {
				tokenID = token.Id
				break
			}
		}
		require.NotEmpty(t, tokenID)

		// Revoke the token
		_, err = c1.RevokeUserAuthToken(ctx, &adminv1.RevokeUserAuthTokenRequest{
			TokenId: tokenID,
		})
		require.NoError(t, err)

		// Check the token is revoked
		res5, err := c1.ListUserAuthTokens(ctx, &adminv1.ListUserAuthTokensRequest{
			UserId: "current",
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(res5.Tokens))
		for _, token := range res5.Tokens {
			require.NotEqual(t, token.Id, tokenID)
		}

	})
}
