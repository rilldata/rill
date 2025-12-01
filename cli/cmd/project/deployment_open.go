package project

import (
	"fmt"

	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeploymentOpenCmd(ch *cmdutil.Helper) *cobra.Command {
	var noOpen bool
	var project, path string

	openCmd := &cobra.Command{
		Use:   "open [<project>] <branch>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Open browser for a specific deployment branch",
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

			// Get the project to retrieve its frontend URL
			projResp, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				Org:     ch.Org,
				Project: project,
			})
			if err != nil {
				return err
			}

			if projResp.Project.FrontendUrl == "" {
				return fmt.Errorf("project does not have a frontend URL")
			}

			// Add branch as a query parameter to the project's frontend URL
			projectURL, err := urlutil.WithQuery(projResp.Project.FrontendUrl, map[string]string{
				"branch": branch,
			})
			if err != nil {
				return err
			}

			if !noOpen {
				ch.Printf("Opening browser at: %s\n", projectURL)
				_ = browser.Open(projectURL)
			} else {
				ch.Printf("Open browser at: %s\n", projectURL)
			}

			return nil
		},
	}
	openCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the browser automatically")
	openCmd.Flags().StringVar(&project, "project", "", "Project name")
	openCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	return openCmd
}
