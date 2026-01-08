package deployment

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, environment string

	listCmd := &cobra.Command{
		Use:   "list [<project>]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "List all deployments for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				project = args[0]
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Get project name from flag or infer it
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			if project == "" {
				return fmt.Errorf("project name is required")
			}

			// fetch the project
			projResp, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				Org:     ch.Org,
				Project: project,
			})
			if err != nil {
				return err
			}

			// fetch the deployments
			req := &adminv1.ListDeploymentsRequest{
				Org:     ch.Org,
				Project: project,
			}
			if environment != "" {
				req.Environment = environment
			}
			resp, err := client.ListDeployments(cmd.Context(), req)
			if err != nil {
				return err
			}

			for _, d := range resp.Deployments {
				if d.Id == projResp.Project.PrimaryDeploymentId {
					d.Branch += " (primary)"
				}
			}
			ch.PrintDeployments(resp.Deployments)
			return nil
		},
	}

	listCmd.Flags().StringVar(&project, "project", "", "Project name")
	listCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	listCmd.Flags().StringVar(&environment, "environment", "", "Filter deployments by environment (prod/dev)")

	return listCmd
}
