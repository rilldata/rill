package deployment

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeploymentCreateCmd(ch *cmdutil.Helper) *cobra.Command {
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

			ch.PrintfSuccess("Deployment created successfully!\n\n")
			ch.Printf("Deployment ID: %s\n", resp.Deployment.Id)
			ch.Printf("Branch: %s\n", resp.Deployment.Branch)
			ch.Printf("Status: %s\n", resp.Deployment.Status.String())
			return nil
		},
	}

	createCmd.Flags().StringVar(&project, "project", "", "Project name")
	createCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	createCmd.Flags().StringVar(&environment, "environment", "dev", "Optional environment to create for (options: dev, prod)")
	createCmd.Flags().BoolVar(&editable, "editable", false, "Make the deployment editable (changes are persisted back to git repo)")

	return createCmd
}
