package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	createCmd := &cobra.Command{
		Use:   "create [<name>]",
		Short: "Create a group",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				name = args[0]
			}

			err = cmdutil.StringPromptIfEmpty(&name, "Enter user group name")
			if err != nil {
				return err
			}

			_, err = client.CreateUsergroup(cmd.Context(), &adminv1.CreateUsergroupRequest{
				Org:  ch.Org,
				Name: name,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q created in organization %q\n", name, ch.Org)

			return nil
		},
	}

	createCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return createCmd
}
