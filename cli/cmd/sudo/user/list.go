package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list <org> [project]",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			orgName := args[0]
			projectName := ""
			if len(args) == 2 {
				projectName = args[1]
			}

			if projectName != "" {
				members, err := client.ListProjectMemberUsers(ctx, &adminv1.ListProjectMemberUsersRequest{
					Org:                  orgName,
					Project:              projectName,
					PageSize:             pageSize,
					PageToken:            pageToken,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return err
				}

				ch.PrintProjectMemberUsers(members.Members)

				if members.NextPageToken != "" {
					cmd.Println()
					cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
				}
			} else {
				members, err := client.ListOrganizationMemberUsers(ctx, &adminv1.ListOrganizationMemberUsersRequest{
					Org:                  orgName,
					PageSize:             pageSize,
					PageToken:            pageToken,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return err
				}

				ch.PrintOrganizationMemberUsers(members.Members)

				if members.NextPageToken != "" {
					cmd.Println()
					cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
				}
			}

			return nil
		},
	}

	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of users to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}
