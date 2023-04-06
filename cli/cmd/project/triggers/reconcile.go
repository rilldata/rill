package triggers

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ReconcileCmd(cfg *config.Config) *cobra.Command {
	reconcileCmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			project, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             args[0],
			})
			if err != nil {
				return err
			}

			// Trigger reconcile (runs in the background - err means the deployment wasn't found, which is unlikely)
			if project.GetProductionDeployment() != nil {
				res, err := client.TriggerReconcile(cmd.Context(), &adminv1.TriggerReconcileRequest{OrganizationName: cfg.Org, Name: args[0]})
				if err != nil {
					return err
				}

				fmt.Println("Reconcile completes", res)
			}

			return nil
		},
	}

	return reconcileCmd
}
