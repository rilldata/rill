package triggers

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RefreshCmd(cfg *config.Config) *cobra.Command {
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh",
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

			// Trigger refresh source (runs in the background - err means the deployment wasn't found, which is unlikely)
			if project.GetProductionDeployment() != nil {
				_, err := client.TriggerRefreshSource(cmd.Context(), &adminv1.TriggerRefreshSourceRequest{OrganizationName: cfg.Org, Name: args[0]})
				if err != nil {
					return err
				}

				fmt.Printf("Refresh source is triggered for project %s, please run 'rill project status` to know the status \n", args[0])
			}

			return nil
		},
	}

	return refreshCmd
}
