package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string
	var description string

	editCmd := &cobra.Command{
		Use:   "edit [<name>]",
		Short: "Edit a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				name = args[0]
			}

			err = cmdutil.StringPromptIfEmpty(&description, "Enter description")
			if err != nil {
				return err
			}

			_, err = client.EditUsergroup(cmd.Context(), &adminv1.EditUsergroupRequest{
				Org:         ch.Org,
				Usergroup:   name,
				Description: description,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q updated\n", name)

			return nil
		},
	}

	editCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	editCmd.Flags().StringVar(&description, "description", "", "Description")

	return editCmd
}
