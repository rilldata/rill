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
	var group string

	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Set role of a user group in an organization or project",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.SelectPromptIfEmpty(&role, "Select role", usergroupRoles, "")
			if err != nil {
				return err
			}

			err = cmdutil.StringPromptIfEmpty(&group, "Enter user group name")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.SetProjectUsergroupRole(cmd.Context(), &adminv1.SetProjectUsergroupRoleRequest{
					Organization: ch.Org,
					Project:      projectName,
					Usergroup:    group,
					Role:         role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user group %q to %q in project %q\n", group, role, projectName)
			} else {
				_, err = client.SetOrganizationUsergroupRole(cmd.Context(), &adminv1.SetOrganizationUsergroupRoleRequest{
					Organization: ch.Org,
					Usergroup:    group,
					Role:         role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user group %q to %q in organization %q\n", group, role, ch.Org)
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&group, "group", "", "User group")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user group (options: %s)", strings.Join(usergroupRoles, ", ")))

	return setRoleCmd
}
