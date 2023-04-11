package project

import (
	"context"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.ListProjectsForOrganization(context.Background(), &adminv1.ListProjectsForOrganizationRequest{
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Projects list \n")
			cmdutil.TablePrinter(toTable(proj.Projects, cfg.Org))
			return nil
		},
	}

	return listCmd
}
