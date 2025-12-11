package deployment

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeploymentStartCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string

	startCmd := &cobra.Command{
		Use:   "start [<project>] <branch>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Start a deployment by branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			var branch string
			if len(args) == 1 {
				branch = args[0]
			} else if len(args) == 2 {
				project = args[0]
				branch = args[1]
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Get project name from flag or infer it
			if !cmd.Flags().Changed("project") && len(args) <= 1 && ch.Interactive {
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			if project == "" {
				return fmt.Errorf("project name is required")
			}

			// List deployments for the project to find the one matching the branch
			resp, err := client.ListDeployments(cmd.Context(), &adminv1.ListDeploymentsRequest{
				Org:     ch.Org,
				Project: project,
				Branch:  branch,
			})
			if err != nil {
				return err
			}

			if len(resp.Deployments) == 0 {
				return fmt.Errorf("no deployment found for branch %q in project %q", branch, project)
			}

			if len(resp.Deployments) > 1 {
				// should not happen in normal circumstances
				return fmt.Errorf("multiple deployments found for branch %q in project %q, cannot proceed. Delete existing deployments first", branch, project)
			}

			deployment := resp.Deployments[0]

			ch.PrintfBold("Starting deployment for branch %q (ID: %s)...\n", branch, deployment.Id)

			startResp, err := client.StartDeployment(cmd.Context(), &adminv1.StartDeploymentRequest{
				DeploymentId: deployment.Id,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Deployment started successfully!\n\n")
			ch.Printf("Deployment ID: %s\n", startResp.Deployment.Id)
			ch.Printf("Branch: %s\n", startResp.Deployment.Branch)
			ch.Printf("Status: %s\n", startResp.Deployment.Status.String())
			return nil
		},
	}

	startCmd.Flags().StringVar(&project, "project", "", "Project name")
	startCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return startCmd
}
