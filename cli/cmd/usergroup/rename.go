package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName string

	createCmd := &cobra.Command{
		Use:   "rename [<name>]",
		Short: "Rename a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				name = args[0]
			}

			err = cmdutil.StringPromptIfEmpty(&newName, "Enter new name")
			if err != nil {
				return err
			}

			_, err = client.RenameUsergroup(cmd.Context(), &adminv1.RenameUsergroupRequest{
				Org:       ch.Org,
				Usergroup: name,
				Name:      newName,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q renamed to %q\n", name, newName)

			return nil
		},
	}

	createCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	createCmd.Flags().StringVar(&newName, "new-name", "", "New user group name")

	return createCmd
}
