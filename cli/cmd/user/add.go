package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(cfg *config.Config) *cobra.Command {
	var orgName string
	var projectName string
	var email string
	var role string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.SelectPromptIfEmpty(&role, "Select role", userRoles, "")
			cmdutil.StringPromptIfEmpty(&email, "Enter email")

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				res, err := client.AddProjectMember(cmd.Context(), &adminv1.AddProjectMemberRequest{
					Organization: orgName,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					cmdutil.SuccessPrinter(fmt.Sprintf("Invitation sent to %q to join project \"%s/%s\" as %q", email, orgName, projectName, role))
				} else {
					cmdutil.SuccessPrinter(fmt.Sprintf("User %q added to the project \"%s/%s\" as %q", email, orgName, projectName, role))
				}
			} else {
				res, err := client.AddOrganizationMember(cmd.Context(), &adminv1.AddOrganizationMemberRequest{
					Organization: orgName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					cmdutil.SuccessPrinter(fmt.Sprintf("Invitation sent to %q to join organization %q as %q", email, orgName, role))
				} else {
					cmdutil.SuccessPrinter(fmt.Sprintf("User %q added to the organization %q as %q", email, orgName, role))
				}
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&orgName, "org", cfg.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user [%v]", strings.Join(userRoles, ", ")))

	return addCmd
}
