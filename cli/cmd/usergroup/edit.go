package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var newName string
	var description string

	editCmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name := args[0]

			req := &adminv1.UpdateUsergroupRequest{
				Org:       ch.Org,
				Usergroup: name,
			}

			if cmd.Flags().Changed("new-name") {
				req.NewName = &newName
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}

			_, err = client.UpdateUsergroup(cmd.Context(), req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User group %q updated\n", name)

			return nil
		},
	}

	editCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	editCmd.Flags().StringVar(&newName, "new-name", "", "New user group name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")

	return editCmd
}
