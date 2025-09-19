package service

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetRoleCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, projectName, role string
	setRoleCmd := &cobra.Command{
		Use:   "set-role <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Set role for service",
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			var roleOptions []string
			if projectName != "" {
				roleOptions = projectRoles
			} else {
				roleOptions = orgRoles
			}

			if ch.Interactive {
				err := cmdutil.StringPromptIfEmpty(&name, "Enter service name")
				if err != nil {
					return err
				}
				err = cmdutil.SelectPromptIfEmpty(&role, "Select role", roleOptions, "")
				if err != nil {
					return err
				}
			} else if name == "" || role == "" {
				return fmt.Errorf("name and role must be specified")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				_, err = client.SetProjectMemberServiceRole(cmd.Context(), &adminv1.SetProjectMemberServiceRoleRequest{
					Name:    name,
					Org:     ch.Org,
					Project: projectName,
					Role:    role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of service %q to %q in the project \"%s/%s\"\n", name, role, ch.Org, projectName)
			} else {
				_, err = client.SetOrganizationMemberServiceRole(cmd.Context(), &adminv1.SetOrganizationMemberServiceRoleRequest{
					Name: name,
					Org:  ch.Org,
					Role: role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Updated role of service %q to %q in the organization \"%s\"\n", name, role, ch.Org)
			}

			return nil
		},
	}
	setRoleCmd.Flags().SortFlags = false
	setRoleCmd.Flags().StringVar(&projectName, "project", "", "Project")
	setRoleCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the service (options: %s)", strings.Join(orgRoles, ", ")))

	return setRoleCmd
}
