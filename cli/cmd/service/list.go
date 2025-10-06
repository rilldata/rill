package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var project string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if project != "" {
				// If project is specified, list services for that project
				res, err := client.ListProjectMemberServices(cmd.Context(), &adminv1.ListProjectMemberServicesRequest{
					Org:     ch.Org,
					Project: project,
				})
				if err != nil {
					return err
				}

				ch.PrintProjectMemberServices(res.Services)
				return nil
			}

			res, err := client.ListServices(cmd.Context(), &adminv1.ListServicesRequest{
				Org: ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintOrganizationMemberServices(res.Services)

			return nil
		},
	}

	listCmd.Flags().StringVar(&project, "project", "", "Project name to filter services")
	return listCmd
}
