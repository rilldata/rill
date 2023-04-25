package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ReconcileCmd(cfg *config.Config) *cobra.Command {
	var name, path, source string
	var refresh, reset bool

	reconcileCmd := &cobra.Command{
		Use:               "reconcile",
		Args:              cobra.NoArgs,
		Hidden:            !cfg.IsDev(),
		Short:             "Send trigger to deployment",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				name, err = inferProjectName(ctx, client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			// Trigger reconcile, refresh, reset (runs in the background - err means the deployment wasn't found, which is unlikely)
			if resp.GetProdDeployment() != nil {
				if refresh {
					fmt.Println("refresh triggered")
					if source != "" {
						_, err := client.TriggerRefreshSource(ctx, &adminv1.TriggerRefreshSourceRequest{OrganizationName: cfg.Org, Name: name, SourceName: source})
						if err != nil {
							return err
						}

						fmt.Printf("Refresh source is triggered for project %s, please run 'rill project status` to know the status \n", name)
						return nil
					}
					return fmt.Errorf("No source name provided")
				}

				if source != "" {
					return fmt.Errorf("`source` flag can only be set with refresh")
				}

				if reset {
					_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{OrganizationName: cfg.Org, Name: name})
					if err != nil {
						return err
					}

					fmt.Printf("Reset project is triggered for project %s, please run 'rill project status` to know the status \n", name)
					return nil
				}

				_, err := client.TriggerReconcile(ctx, &adminv1.TriggerReconcileRequest{OrganizationName: cfg.Org, Name: name})
				if err != nil {
					return err
				}

				fmt.Printf("Reconcile is triggered for project %s, please run 'rill project status` to know the status \n", name)
			}

			return nil
		},
	}

	reconcileCmd.Flags().SortFlags = false
	reconcileCmd.Flags().StringVar(&name, "project", "", "Name")
	reconcileCmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh")
	reconcileCmd.Flags().BoolVar(&reset, "reset", false, "Reset")
	reconcileCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	reconcileCmd.Flags().StringVar(&source, "source", "", "Source Name")

	return reconcileCmd
}
