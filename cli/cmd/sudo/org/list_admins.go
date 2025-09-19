package org

import (
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListAdminsCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-admins <org>",
		Args:  cobra.ExactArgs(1),
		Short: "Show all admin users for an org",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var members []*adminv1.OrganizationMemberUser
			pageSize := uint32(1000)
			pageToken := ""
			for {
				resp, err := client.ListOrganizationMemberUsers(cmd.Context(), &adminv1.ListOrganizationMemberUsersRequest{
					Org:                  args[0],
					PageSize:             pageSize,
					PageToken:            pageToken,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return err
				}

				for _, m := range resp.Members {
					if m.RoleName == database.OrganizationRoleNameAdmin {
						members = append(members, m)
					}
				}

				pageToken = resp.NextPageToken
				if pageToken == "" {
					break
				}
			}

			if len(members) == 0 {
				ch.PrintfError("No admin users found for org %q. This should not be possible.\n", args[0])
				return nil
			}

			ch.PrintOrganizationMemberUsers(members)

			return nil
		},
	}

	return cmd
}
