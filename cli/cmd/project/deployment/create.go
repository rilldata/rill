package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, environment string
	var editable bool

	createCmd := &cobra.Command{
		Use:   "create [<project>] <branch>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Create a deployment for a specific branch",
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

			if environment == "prod" && editable {
				return fmt.Errorf("prod deployments cannot be editable")
			}

			ch.PrintfBold("Creating %q deployment for branch %q...\n", environment, branch)

			resp, err := client.CreateDeployment(cmd.Context(), &adminv1.CreateDeploymentRequest{
				Org:         ch.Org,
				Project:     project,
				Environment: environment,
				Branch:      branch,
				Editable:    editable,
			})
			if err != nil {
				return err
			}

			ch.Printf("Deployment created with branch %q\n", resp.Deployment.Branch)
			ch.Printf("Provisioning runtime (this may take a moment)")

			// Poll for deployment status and print result
			return pollDeploymentStatus(cmd.Context(), client, ch, resp.Deployment.Id, project, branch, environment)
		},
	}

	createCmd.Flags().StringVar(&project, "project", "", "Project name")
	createCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	createCmd.Flags().StringVar(&environment, "environment", "dev", "Optional environment to create for (options: dev, prod)")
	createCmd.Flags().BoolVar(&editable, "editable", false, "Make the deployment editable (changes are persisted back to git repo)")

	return createCmd
}

// pollDeploymentStatus polls the deployment status until it's either running or errored, then prints the result
func pollDeploymentStatus(ctx context.Context, client adminv1.AdminServiceClient, ch *cmdutil.Helper, deploymentID, project, branch, environment string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ch.Printf(".")

			// Fetch the deployment status using ListDeployments filtered by branch
			resp, err := client.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
				Org:         ch.Org,
				Project:     project,
				Environment: environment,
				Branch:      branch,
			})
			if err != nil {
				return err
			}

			// Find the deployment with matching ID - usually there's only one per branch
			var deployment *adminv1.Deployment
			for _, d := range resp.Deployments {
				if d.Id == deploymentID {
					deployment = d
					break
				}
			}

			if deployment == nil {
				return fmt.Errorf("deployment not found")
			}

			// Check if deployment is in a terminal state and print result
			switch deployment.Status {
			case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING:
				ch.PrintfSuccess("\n\nRuntime provisioned successfully!\n\n")
				ch.Printf("Runtime Host: %s\n", deployment.RuntimeHost)
				return nil
			case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERRORED:
				ch.Printf("\n\n")
				return fmt.Errorf("runtime provisioning failed: %s", deployment.StatusMessage)
			}
			// Continue polling for other statuses (PENDING, UPDATING, etc.)
		}
	}
}
