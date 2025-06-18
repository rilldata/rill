package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string

	deleteCmd := &cobra.Command{
		Use:   "delete <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if projectName != "" {
				// If projectName is provided, delete the service from the specified project
				_, err = client.RemoveProjectMemberService(cmd.Context(), &adminv1.RemoveProjectMemberServiceRequest{
					Name:             args[0],
					OrganizationName: ch.Org,
					ProjectName:      projectName,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Deleted service: %q from project: %q\n", args[0], projectName)
				return nil
			}

			_, err = client.DeleteService(cmd.Context(), &adminv1.DeleteServiceRequest{
				Name:             args[0],
				OrganizationName: ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Deleted service: %q\n", args[0])

			return nil
		},
	}

	deleteCmd.Flags().StringVar(&projectName, "project", "", "Project to remove service from")
	return deleteCmd
}
