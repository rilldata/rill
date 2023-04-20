package user

import (
	"errors"
	"fmt"

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
			if orgName == "" && projectName == "" {
				return errors.New("either organization or project has to be specified")
			}
			if orgName != "" && projectName != "" {
				return errors.New("only one of organization or project has to be specified")
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			var pendingSignup bool
			if orgName != "" {
				res, err := client.AddOrganizationMember(cmd.Context(), &adminv1.AddOrganizationMemberRequest{
					Organization: orgName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				pendingSignup = res.PendingSignup
			} else {
				res, err := client.AddProjectMember(cmd.Context(), &adminv1.AddProjectMemberRequest{
					Organization: cfg.Org,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}
				pendingSignup = res.PendingSignup
			}

			if pendingSignup {
				cmdutil.SuccessPrinter(fmt.Sprintf("Invitation sent to %q to join project %q as %q", email, projectName, role))
				return nil
			}
			cmdutil.SuccessPrinter(fmt.Sprintf("User %q added to the project %q as %q", email, projectName, role))
			return nil
		},
	}

	addCmd.Flags().StringVar(&orgName, "org", "", "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", "Role of the user, should be admin/viewer")

	return addCmd
}
