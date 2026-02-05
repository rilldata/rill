package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var groupName string
	var email string

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.StringPromptIfEmpty(&email, "Enter email")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if groupName != "" {
				_, err = client.RemoveUsergroupMemberUser(cmd.Context(), &adminv1.RemoveUsergroupMemberUserRequest{
					Org:       ch.Org,
					Usergroup: groupName,
					Email:     email,
				})
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Removed user %q from user group \"%s/%s\"\n", email, ch.Org, groupName)
				return nil
			} else if projectName != "" {
				_, err = client.RemoveProjectMemberUser(cmd.Context(), &adminv1.RemoveProjectMemberUserRequest{
					Org:     ch.Org,
					Project: projectName,
					Email:   email,
				})
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Removed user %q from project \"%s/%s\"\n", email, ch.Org, projectName)
			} else {
				_, err = client.RemoveOrganizationMemberUser(cmd.Context(), &adminv1.RemoveOrganizationMemberUserRequest{
					Org:   ch.Org,
					Email: email,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Removed user %q from organization %q\n", email, ch.Org)
			}

			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	removeCmd.Flags().StringVar(&projectName, "project", "", "Project")
	removeCmd.Flags().StringVar(&groupName, "group", "", "User group")
	removeCmd.Flags().StringVar(&email, "email", "", "Email of the user")

	return removeCmd
}
