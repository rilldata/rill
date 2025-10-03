package org

import (
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func JoinCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join <org>",
		Args:  cobra.ExactArgs(1),
		Short: "Add yourself as a permanent admin member of an org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ch.PrintfWarn("This command will permanently add you as an admin member of %q and your name will show up in member listings. ", args[0])
			ch.PrintfWarn("If you only need temporary access, consider instead assuming the identity of an existing admin using `rill sudo org list-admins` and `rill sudo user assume`.\n")
			ok, err := cmdutil.ConfirmPrompt("Do you want to proceed?", "", true)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted.\n")
				return nil
			}

			c, err := ch.Client()
			if err != nil {
				return err
			}

			me, err := c.GetCurrentUser(cmd.Context(), &adminv1.GetCurrentUserRequest{})
			if err != nil {
				return err
			}

			_, err = c.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
				Org:                  args[0],
				Email:                me.User.Email,
				Role:                 database.OrganizationRoleNameAdmin,
				SuperuserForceAccess: true,
			})
			if err != nil {
				// Optimistically retry if the user is already a member (but might not be an admin).
				_, retryErr := c.SetOrganizationMemberUserRole(cmd.Context(), &adminv1.SetOrganizationMemberUserRoleRequest{
					Org:                  args[0],
					Email:                me.User.Email,
					Role:                 database.OrganizationRoleNameAdmin,
					SuperuserForceAccess: true,
				})
				if retryErr != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}
