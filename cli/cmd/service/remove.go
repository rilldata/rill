package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string

	removeCmd := &cobra.Command{
		Use:   "remove <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "remove service from org or project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				// If projectName is provided, delete the service from the specified project
				_, err = client.RemoveProjectMemberService(cmd.Context(), &adminv1.RemoveProjectMemberServiceRequest{
					Name:    args[0],
					Org:     ch.Org,
					Project: projectName,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Removed service %q from project %q\n", args[0], projectName)
				return nil
			}

			_, err = client.RemoveOrganizationMemberService(cmd.Context(), &adminv1.RemoveOrganizationMemberServiceRequest{
				Name: args[0],
				Org:  ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Removed service %q from org %q\n", args[0], ch.Org)

			return nil
		},
	}

	removeCmd.Flags().StringVar(&projectName, "project", "", "Project to remove service from")
	return removeCmd
}
