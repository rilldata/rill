package server_test

import (
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOrganizationDefaultProvisioner(t *testing.T) {
	fix := testadmin.New(t)
	_, sc := fix.NewSuperuser(t)
	_, uc := fix.NewUser(t)

	// The user creates an org, which makes them an org admin.
	org, err := uc.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	require.Empty(t, org.Organization.DefaultProvisioner)

	t.Run("org admin cannot set the default provisioner", func(t *testing.T) {
		// UpdateOrganization no longer exposes the default provisioner, so an org admin cannot change it there.
		// A plain update must not touch the stored value.
		desc := "updated description"
		res, err := uc.UpdateOrganization(t.Context(), &adminv1.UpdateOrganizationRequest{
			Org:         org.Organization.Name,
			Description: &desc,
		})
		require.NoError(t, err)
		require.Equal(t, desc, res.Organization.Description)
		require.Empty(t, res.Organization.DefaultProvisioner)

		// The sudo RPC is rejected for non-superusers, even for an org admin.
		_, err = uc.SudoUpdateOrganizationDefaultProvisioner(t.Context(), &adminv1.SudoUpdateOrganizationDefaultProvisionerRequest{
			Org:                org.Organization.Name,
			DefaultProvisioner: "static",
		})
		require.Error(t, err)
		require.Equal(t, codes.PermissionDenied, status.Code(err))

		shown, err := uc.GetOrganization(t.Context(), &adminv1.GetOrganizationRequest{Org: org.Organization.Name})
		require.NoError(t, err)
		require.Empty(t, shown.Organization.DefaultProvisioner)
	})

	t.Run("superuser can set the default provisioner via the sudo RPC", func(t *testing.T) {
		// "static" is the provisioner configured by testadmin.
		res, err := sc.SudoUpdateOrganizationDefaultProvisioner(t.Context(), &adminv1.SudoUpdateOrganizationDefaultProvisionerRequest{
			Org:                org.Organization.Name,
			DefaultProvisioner: "static",
		})
		require.NoError(t, err)
		require.Equal(t, "static", res.Organization.DefaultProvisioner)

		// The value is visible to the org admin, and an ordinary update by the org admin does not clear it.
		desc := "another description"
		updated, err := uc.UpdateOrganization(t.Context(), &adminv1.UpdateOrganizationRequest{
			Org:         org.Organization.Name,
			Description: &desc,
		})
		require.NoError(t, err)
		require.Equal(t, "static", updated.Organization.DefaultProvisioner)

		// An unknown provisioner is rejected.
		_, err = sc.SudoUpdateOrganizationDefaultProvisioner(t.Context(), &adminv1.SudoUpdateOrganizationDefaultProvisionerRequest{
			Org:                org.Organization.Name,
			DefaultProvisioner: "nonexistent",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), `provisioner "nonexistent" is not configured`)

		// The superuser can clear it again.
		res, err = sc.SudoUpdateOrganizationDefaultProvisioner(t.Context(), &adminv1.SudoUpdateOrganizationDefaultProvisionerRequest{
			Org:                org.Organization.Name,
			DefaultProvisioner: "",
		})
		require.NoError(t, err)
		require.Empty(t, res.Organization.DefaultProvisioner)
	})
}
