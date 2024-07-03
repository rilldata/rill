package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveUserCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string
	var group string

	removeCmd := &cobra.Command{
		Use:   "remove-user <email>",
		Short: "Remove a user from a user group",
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

			_, err = client.RemoveUsergroupMember(cmd.Context(), &adminv1.RemoveUsergroupMemberRequest{
				Organization: ch.Org,
				Usergroup:    group,
				Email:        email,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User %q removed from the user group %q\n", email, group)

			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	removeCmd.Flags().StringVar(&group, "group", "", "Name of the user group")

	return removeCmd

}
