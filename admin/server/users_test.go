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
		_, err = c2.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, grpc.Code(err))

		// A superuser can delete any user
		_, err = sc1.DeleteUser(ctx, &adminv1.DeleteUserRequest{
			Email: u3.Email,
		})
		require.NoError(t, err)
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
}
