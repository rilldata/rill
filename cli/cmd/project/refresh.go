package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RefreshCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var source []string
	cfg := ch.Config

	refreshCmd := &cobra.Command{
		Use:               "refresh [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Refresh the project's data sources",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				project = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				var err error
				project, err = inferProjectName(ctx, client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             project,
			})
			if err != nil {
				return err
			}

			if resp.ProdDeployment == nil {
				return fmt.Errorf("no production deployment found for project %q", project)
			}

			_, err = client.TriggerRefreshSources(ctx, &adminv1.TriggerRefreshSourcesRequest{DeploymentId: resp.ProdDeployment.Id, Sources: source})
			if err != nil {
				return fmt.Errorf("failed to trigger refresh: %w", err)
			}

			fmt.Printf("Triggered refresh. To see status, run `rill project status --project %s`.\n", project)

			return nil
		},
	}

	refreshCmd.Flags().SortFlags = false
	refreshCmd.Flags().StringVar(&project, "project", "", "Project name")
	refreshCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	refreshCmd.Flags().StringSliceVar(&source, "source", nil, "Refresh specific source(s)")

	return refreshCmd
}
