package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var groupName string

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a group's role on a project or organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.StringPromptIfEmpty(&groupName, "Enter user group name")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.RemoveProjectMemberUsergroup(cmd.Context(), &adminv1.RemoveProjectMemberUsergroupRequest{
					Org:       ch.Org,
					Project:   projectName,
					Usergroup: groupName,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Removed role of user group %q in the project %q\n", groupName, projectName)
			} else {
				_, err = client.RemoveOrganizationMemberUsergroup(cmd.Context(), &adminv1.RemoveOrganizationMemberUsergroupRequest{
					Org:       ch.Org,
					Usergroup: groupName,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Removed role of user group %q in the organization %q\n", groupName, ch.Org)
			}

			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	removeCmd.Flags().StringVar(&projectName, "project", "", "Project")
	removeCmd.Flags().StringVar(&groupName, "group", "", "User group")

	return removeCmd
}
