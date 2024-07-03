package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a user group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name = args[0]

			_, err = client.RemoveUsergroup(cmd.Context(), &adminv1.RemoveUsergroupRequest{
				Organization: ch.Org,
				Usergroup:    name,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q of organization %q removed\n", name, ch.Org)

			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return removeCmd
}
