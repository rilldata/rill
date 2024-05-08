package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ReconcileCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var refresh, reset, force bool
	var refreshSources []string

	reconcileCmd := &cobra.Command{
		Use:               "reconcile [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Send trigger to deployment",
		Hidden:            true,
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				project = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(ctx, ch.Org, path)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: ch.Org,
				Name:             project,
			})
			if err != nil {
				return err
			}

			if reset || resp.ProdDeployment == nil {
				if !force {
					msg := "This will create a new deployment, causing downtime as data sources are reloaded from scratch. If you just need to refresh data, use `rill project refresh`. Do you want to continue?"
					ok, err := cmdutil.ConfirmPrompt(msg, "", false)
					if err != nil {
						return err
					}
					if !ok {
						return nil
					}
				}

				_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{Organization: ch.Org, Project: project})
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
	reconcileCmd.Flags().BoolVar(&force, "force", false, "Force the operation")

	reconcileCmd.MarkFlagsMutuallyExclusive("reset", "refresh")
	reconcileCmd.MarkFlagsMutuallyExclusive("reset", "refresh-source")

	return reconcileCmd
}
