package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetRoleCmd(cfg *config.Config) *cobra.Command {
	var orgName string
	var projectName string
	var email string
	var role string

	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Set Role",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgName == "" {
				orgName = cfg.Org
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			cmdutil.SelectPromptIfEmpty(&role, "Please enter the user's role.", []string{"admin", "viewer"}, "")
			cmdutil.StringPromptIfEmpty(&email, "Please enter the email of the user.")

			if projectName != "" {
				_, err = client.SetProjectMemberRole(cmd.Context(), &adminv1.SetProjectMemberRoleRequest{
					Organization: cfg.Org,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				cmdutil.SuccessPrinter(fmt.Sprintf("Updated role of user %q to %q in the project %q under organization %q", email, role, projectName, orgName))
			} else {
				_, err = client.SetOrganizationMemberRole(cmd.Context(), &adminv1.SetOrganizationMemberRoleRequest{
					Organization: orgName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				cmdutil.SuccessPrinter(fmt.Sprintf("Updated role of user %q to %q in the organization %q", email, role, orgName))
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&orgName, "org", "", "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	setRoleCmd.Flags().StringVar(&role, "role", "", "Role of the user, should be admin/viewer")

	return setRoleCmd
}
