package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddUserCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string
	var group string

	addCmd := &cobra.Command{
		Use:   "add-user <email>",
		Short: "Add a user to a user group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			email = args[0]

			err = cmdutil.StringPromptIfEmpty(&group, "Enter user group name")
			if err != nil {
				return err
			}

			_, err = client.AddUsergroupMember(cmd.Context(), &adminv1.AddUsergroupMemberRequest{
				Organization: ch.Org,
				Usergroup:    group,
				Email:        email,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User %q added to the user group %q\n", email, group)

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&group, "group", "", "Name of the user group")

	return addCmd
}
