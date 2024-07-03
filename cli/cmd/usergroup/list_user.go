package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListUserCmd(ch *cmdutil.Helper) *cobra.Command {
	var group string

	addCmd := &cobra.Command{
		Use:   "list-user <name>",
		Short: "List users in a user group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			group = args[0]

			res, err := client.ListUsergroupMembers(cmd.Context(), &adminv1.ListUsergroupMembersRequest{
				Organization: ch.Org,
				Usergroup:    group,
			})
			if err != nil {
				return err
			}

			ch.PrintUsergroupMembers(res.Members)

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return addCmd
}
