package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var role string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.SelectPromptIfEmpty(&role, "Select role", userRoles, "")
			cmdutil.StringPromptIfEmpty(&email, "Enter email")

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				res, err := client.AddProjectMember(cmd.Context(), &adminv1.AddProjectMemberRequest{
					Organization: ch.Org,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				} else {
					ch.PrintfSuccess("User %q added to the project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				}
			} else {
				res, err := client.AddOrganizationMember(cmd.Context(), &adminv1.AddOrganizationMemberRequest{
					Organization: ch.Org,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q\n", email, ch.Org, role)
				} else {
					ch.PrintfSuccess("User %q added to the organization %q as %q\n", email, ch.Org, role)
				}
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user [%v]", strings.Join(userRoles, ", ")))

	return addCmd
}
