package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetRoleCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	var email string
	var role string

	setRoleCmd := &cobra.Command{
		Use:   "set-role",
		Short: "Set Role",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.SelectPromptIfEmpty(&role, "Select role", userRoles, "")
			cmdutil.StringPromptIfEmpty(&email, "Enter email")

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				_, err = client.SetProjectMemberRole(cmd.Context(), &adminv1.SetProjectMemberRoleRequest{
					Organization: orgName,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				cmdutil.SuccessPrinter(fmt.Sprintf("Updated role of user %q to %q in the project \"%s/%s\"", email, role, cfg.Org, projectName))
			} else {
				_, err = client.SetOrganizationMemberRole(cmd.Context(), &adminv1.SetOrganizationMemberRoleRequest{
					Organization: cfg.Org,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				cmdutil.SuccessPrinter(fmt.Sprintf("Updated role of user %q to %q in the organization %q", email, role, cfg.Org))
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user [%v]", strings.Join(userRoles, ", ")))

	return setRoleCmd
}
