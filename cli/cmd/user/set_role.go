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
		Short: "Set Role",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.SelectPromptIfEmpty(&role, "Select role", userRoles, "")
			cmdutil.StringPromptIfEmpty(&email, "Enter email")

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.SetProjectMemberRole(cmd.Context(), &adminv1.SetProjectMemberRoleRequest{
					Organization: ch.Org,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				ch.Printer.PrintlnSuccess(fmt.Sprintf("Updated role of user %q to %q in the project \"%s/%s\"", email, role, ch.Org, projectName))
			} else {
				_, err = client.SetOrganizationMemberRole(cmd.Context(), &adminv1.SetOrganizationMemberRoleRequest{
					Organization: ch.Org,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				ch.Printer.PrintlnSuccess(fmt.Sprintf("Updated role of user %q to %q in the organization %q", email, role, ch.Org))
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user [%v]", strings.Join(userRoles, ", ")))

	return setRoleCmd
}
