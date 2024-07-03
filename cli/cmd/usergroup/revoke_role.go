package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RevokeRoleCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var group string

	setRoleCmd := &cobra.Command{
		Use:   "revoke-role",
		Short: "Revoke role of a user group in an organization or project",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmdutil.StringPromptIfEmpty(&group, "Enter user group name")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.RevokeProjectUsergroupRole(cmd.Context(), &adminv1.RevokeProjectUsergroupRoleRequest{
					Organization: ch.Org,
					Project:      projectName,
					Usergroup:    group,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Revoked role of user group %q in the project %q\n", group, projectName)
			} else {
				_, err = client.RevokeOrganizationUsergroupRole(cmd.Context(), &adminv1.RevokeOrganizationUsergroupRoleRequest{
					Organization: ch.Org,
					Usergroup:    group,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Revoked role of user group %q in the organization %q\n", group, ch.Org)
			}

			return nil
		},
	}

	setRoleCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&group, "group", "", "Name of the user group")

	return setRoleCmd
}
