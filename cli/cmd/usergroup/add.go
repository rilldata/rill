package usergroup

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var role string
	var groupName string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a group to a project or organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.SelectPromptIfEmpty(&role, "Select role", usergroupRoles, "")
			if err != nil {
				return err
			}

			err = cmdutil.StringPromptIfEmpty(&groupName, "Enter user group name")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.AddProjectMemberUsergroup(cmd.Context(), &adminv1.AddProjectMemberUsergroupRequest{
					Org:       ch.Org,
					Project:   projectName,
					Usergroup: groupName,
					Role:      role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Role %q added to user group %q in project %q\n", role, groupName, projectName)
			} else {
				_, err = client.AddOrganizationMemberUsergroup(cmd.Context(), &adminv1.AddOrganizationMemberUsergroupRequest{
					Org:       ch.Org,
					Usergroup: groupName,
					Role:      role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Role %q added to user group %q in organization %q\n", role, groupName, ch.Org)
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&groupName, "group", "", "User group")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user group (options: %s)", strings.Join(usergroupRoles, ", ")))

	return addCmd
}
