package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// restartStatusTimeout is the max time to wait for a deployment to stop or start.
const restartStatusTimeout = 20 * time.Minute

func RestartCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	restartCmd := &cobra.Command{
		Use:   "restart <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Restart the project's deployments",
		Long: `Stop and then start each running deployment of the project.
Unlike "reset", this keeps the existing deployments, so the deployment and instance IDs stay the same.
Use it to apply a provisioner change to running deployments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			org := args[0]
			project := args[1]

			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
				Org:                  org,
				Project:              project,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			// Only restart running deployments.
			// Deployments in other states pick up changes the next time they start.
			var running []*adminv1.Deployment
			for _, d := range resp.Deployments {
				if d.Status == adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
					running = append(running, d)
					continue
				}
				ch.Printf("Skipping deployment %s (branch %q) with status %s; it will pick up changes the next time it starts.\n", d.Id, d.Branch, d.Status)
			}
			if len(running) == 0 {
				ch.Printf("No running deployments to restart.\n")
				return nil
			}

			if !force {
				ch.PrintfWarn("%d deployment(s) will be stopped and started one at a time. Each will be unavailable while it restarts.\n", len(running))
				if !ch.Interactive {
					return fmt.Errorf("confirmation required; use --force flag to proceed")
				}
				if err := cmdutil.ConfirmPrompt("Continue?", false); err != nil {
					return err
				}
			}

			for _, d := range running {
				ch.PrintfBold("Restarting deployment %s (branch %q)...\n", d.Id, d.Branch)

				_, err = client.StopDeployment(ctx, &adminv1.StopDeploymentRequest{
					DeploymentId:         d.Id,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return fmt.Errorf("failed to stop deployment %s: %w", d.Id, err)
				}

				// We must wait for the deployment to stop before starting it.
				// Otherwise the reconcile job sees a running deployment that should be running, and doesn't restart it.
				ch.Printf("Stopping...\n")
				_, err = awaitDeploymentStatus(ctx, client, org, project, d, adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPED)
				if err != nil {
					return err
				}

				_, err = client.StartDeployment(ctx, &adminv1.StartDeploymentRequest{
					DeploymentId:         d.Id,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return fmt.Errorf("failed to start deployment %s: %w", d.Id, err)
				}

				ch.Printf("Starting...\n")
				started, err := awaitDeploymentStatus(ctx, client, org, project, d, adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING)
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Restarted deployment %s (runtime host: %s -> %s)\n", d.Id, d.RuntimeHost, started.RuntimeHost)
			}

			return nil
		},
	}

	restartCmd.Flags().SortFlags = false
	restartCmd.Flags().BoolVar(&force, "force", false, "Skip the confirmation prompt")
	return restartCmd
}

// awaitDeploymentStatus polls the deployment until it reaches the target status.
// It returns an error if the deployment errors, disappears, or doesn't reach the target status within restartStatusTimeout.
func awaitDeploymentStatus(ctx context.Context, client adminv1.AdminServiceClient, org, project string, depl *adminv1.Deployment, target adminv1.DeploymentStatus) (*adminv1.Deployment, error) {
	ctx, cancel := context.WithTimeout(ctx, restartStatusTimeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var last *adminv1.Deployment
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) && last != nil {
				return nil, fmt.Errorf("timed out waiting for deployment %s to reach status %s (current status: %s, message: %q)", depl.Id, target, last.Status, last.StatusMessage)
			}
			return nil, ctx.Err()
		case <-ticker.C:
		}

		resp, err := client.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
			Org:                  org,
			Project:              project,
			Environment:          depl.Environment,
			Branch:               depl.Branch,
			SuperuserForceAccess: true,
		})
		if err != nil {
			return nil, err
		}

		last = nil
		for _, d := range resp.Deployments {
			if d.Id == depl.Id {
				last = d
				break
			}
		}
		if last == nil {
			return nil, fmt.Errorf("deployment %s not found", depl.Id)
		}

		switch last.Status {
		case target:
			return last, nil
		case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERRORED:
			return nil, fmt.Errorf("deployment %s errored: %s", depl.Id, last.StatusMessage)
		}
	}
}
