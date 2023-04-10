package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func StatusCmd(cfg *config.Config) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status <project-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             args[0],
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Found project\n")
			cmdutil.TablePrinter(toRow(proj.Project))

			depl := proj.ProductionDeployment
			if depl != nil {
				cmdutil.SuccessPrinter("Deployment info\n")
				fmt.Printf("  Runtime: %s\n", depl.RuntimeHost)
				fmt.Printf("  Instance: %s\n", depl.RuntimeInstanceId)
				fmt.Printf("  Slots: %d\n", depl.Slots)
				fmt.Printf("  Branch: %s\n", depl.Branch)
				fmt.Printf("  Status: %s\n", depl.Status.String())
				fmt.Printf("  Logs: %s\n\n", depl.Logs)
			}

			return nil
		},
	}

	return statusCmd
}
