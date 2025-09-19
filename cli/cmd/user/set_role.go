package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetRoleCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var role string

	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Change a user's role",
		RunE: func(cmd *cobra.Command, args []string) error {
			var roleOptions []string
			if projectName != "" {
				roleOptions = projectRoles
			} else {
				roleOptions = orgRoles
			}

			if ch.Interactive {
				err := cmdutil.SelectPromptIfEmpty(&role, "Select role", roleOptions, "")
				if err != nil {
					return err
				}

				err = cmdutil.StringPromptIfEmpty(&email, "Enter email")
				if err != nil {
					return err
				}
			} else if email == "" || role == "" {
				return fmt.Errorf("email and role must be specified")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.SetProjectMemberUserRole(cmd.Context(), &adminv1.SetProjectMemberUserRoleRequest{
					Org:     ch.Org,
					Project: projectName,
					Email:   email,
					Role:    role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user %q to %q in the project \"%s/%s\"\n", email, role, ch.Org, projectName)
			} else {
				_, err = client.SetOrganizationMemberUserRole(cmd.Context(), &adminv1.SetOrganizationMemberUserRoleRequest{
					Org:   ch.Org,
					Email: email,
					Role:  role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of user %q to %q in the organization %q\n", email, role, ch.Org)
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user (options: %s)", strings.Join(orgRoles, ", ")))

	return setRoleCmd
}
