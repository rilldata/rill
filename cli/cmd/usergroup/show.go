package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	showCmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a user group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name = args[0]

			res, err := client.GetUsergroup(cmd.Context(), &adminv1.GetUsergroupRequest{
				Organization: ch.Org,
				Usergroup:    name,
			})
			if err != nil {
				return err
			}

			ch.PrintUsergroups([]*adminv1.Usergroup{res.Usergroup})

			return nil
		},
	}

	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return showCmd
}
