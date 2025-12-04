package usergroup

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetRoleCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var role string
	var groupName string

	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Change a group's role on a project or organization",
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
				_, err = client.SetProjectMemberUsergroupRole(cmd.Context(), &adminv1.SetProjectMemberUsergroupRoleRequest{
					Org:       ch.Org,
					Project:   projectName,
					Usergroup: groupName,
					Role:      role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user group %q to %q in project %q\n", groupName, role, projectName)
			} else {
				_, err = client.SetOrganizationMemberUsergroupRole(cmd.Context(), &adminv1.SetOrganizationMemberUsergroupRoleRequest{
					Org:       ch.Org,
					Usergroup: groupName,
					Role:      role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user group %q to %q in organization %q\n", groupName, role, ch.Org)
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&groupName, "group", "", "User group")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user group (options: %s)", strings.Join(usergroupRoles, ", ")))

	return setRoleCmd
}
