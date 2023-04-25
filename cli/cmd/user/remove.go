package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(cfg *config.Config) *cobra.Command {
	var orgName string
	var projectName string
	var email string
	var removeFromProjects bool

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.StringPromptIfEmpty(&email, "Enter email")

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				_, err = client.RemoveProjectMember(cmd.Context(), &adminv1.RemoveProjectMemberRequest{
					Organization: orgName,
					Project:      projectName,
					Email:        email,
				})
				if err != nil {
					return err
				}

				cmdutil.SuccessPrinter(fmt.Sprintf("Removed user %q from project \"%s/%s\"", email, orgName, projectName))
			} else {
				_, err = client.RemoveOrganizationMember(cmd.Context(), &adminv1.RemoveOrganizationMemberRequest{
					Organization:       orgName,
					Email:              email,
					RemoveFromProjects: removeFromProjects,
				})
				if err != nil {
					return err
				}
				cmdutil.SuccessPrinter(fmt.Sprintf("Removed user %q from organization %q", email, orgName))
			}

			return nil
		},
	}

	removeCmd.Flags().StringVar(&orgName, "org", cfg.Org, "Organization")
	removeCmd.Flags().StringVar(&projectName, "project", "", "Project")
	removeCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	removeCmd.Flags().BoolVar(&removeFromProjects, "remove-from-projects", false, "Remove user from projects")

	return removeCmd
}
