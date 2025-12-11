package deployment

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeploymentShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string

	showCmd := &cobra.Command{
		Use:   "show [<project>] <branch>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Show deployment details by branch",
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

			ch.PrintDeployments(resp.Deployments)
			return nil
		},
	}

	showCmd.Flags().StringVar(&project, "project", "", "Project name")
	showCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return showCmd
}
