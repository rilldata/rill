package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	deleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name = args[0]

			_, err = client.DeleteUsergroup(cmd.Context(), &adminv1.DeleteUsergroupRequest{
				Org:       ch.Org,
				Usergroup: name,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q of organization %q deleted\n", name, ch.Org)

			return nil
		},
	}

	deleteCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return deleteCmd
}
