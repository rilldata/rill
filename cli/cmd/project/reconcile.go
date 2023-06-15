package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ReconcileCmd(cfg *config.Config) *cobra.Command {
	var project, path string
	var refresh, reset bool
	var refreshSources []string

	reconcileCmd := &cobra.Command{
		Use:               "reconcile [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Send trigger to deployment",
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
				cmdutil.PrintlnWarn("Project does not have a production deployment")
				return nil
			}

			if reset {
				_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{DeploymentId: resp.ProdDeployment.Id})
				if err != nil {
					return err
				}

				fmt.Printf("Triggered project reset. To see status, run `rill project status --project %s`.\n", project)
				return nil
			}

			if refresh || len(refreshSources) > 0 {
				_, err := client.TriggerRefreshSources(ctx, &adminv1.TriggerRefreshSourcesRequest{DeploymentId: resp.ProdDeployment.Id, Sources: refreshSources})
				if err != nil {
					return err
				}

				fmt.Printf("Triggered refresh. To see status, run `rill project status --project %s`.\n", project)
				return nil
			}

			// When neither --reset nor --refresh/--refresh-source is specified, trigger reconcile.
			_, err = client.TriggerReconcile(ctx, &adminv1.TriggerReconcileRequest{DeploymentId: resp.ProdDeployment.Id})
			if err != nil {
				return err
			}

			fmt.Printf("Triggered reconcile. To see status, run `rill project status --project %s`.\n", project)
			return nil
		},
	}

	reconcileCmd.Flags().SortFlags = false
	reconcileCmd.Flags().StringVar(&project, "project", "", "Project name")
	reconcileCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	reconcileCmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh all sources")
	reconcileCmd.Flags().StringSliceVar(&refreshSources, "refresh-source", nil, "Refresh specific source(s)")
	reconcileCmd.Flags().BoolVar(&reset, "reset", false, "Reset and redeploy the project from scratch")

	reconcileCmd.MarkFlagsMutuallyExclusive("reset", "refresh")
	reconcileCmd.MarkFlagsMutuallyExclusive("reset", "refresh-source")

	return reconcileCmd
}
