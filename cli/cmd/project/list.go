package project

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all the projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListProjectsForOrganization(context.Background(), &adminv1.ListProjectsForOrganizationRequest{
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			if len(res.Projects) == 0 {
				cmdutil.WarnPrinter("No projects found")
				return nil
			}

			cmdutil.SuccessPrinter("Projects list")
			cmdutil.TablePrinter(toTable(res.Projects))

			return nil
		},
	}

	return listCmd
}
