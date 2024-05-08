package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var keepProjectRoles bool

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.StringPromptIfEmpty(&email, "Enter email")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.RemoveProjectMember(cmd.Context(), &adminv1.RemoveProjectMemberRequest{
					Organization: ch.Org,
					Project:      projectName,
					Email:        email,
				})
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Removed user %q from project \"%s/%s\"\n", email, ch.Org, projectName)
			} else {
				_, err = client.RemoveOrganizationMember(cmd.Context(), &adminv1.RemoveOrganizationMemberRequest{
					Organization:     ch.Org,
					Email:            email,
					KeepProjectRoles: keepProjectRoles,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Removed user %q from organization %q\n", email, ch.Org)
			}

			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	removeCmd.Flags().StringVar(&projectName, "project", "", "Project")
	removeCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	removeCmd.Flags().BoolVar(&keepProjectRoles, "keep-project-roles", false, "Keep roles granted directly on projects in the org")

	return removeCmd
}
