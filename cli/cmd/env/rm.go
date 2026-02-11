package env

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// RmCmd is sub command for env. Removes the variable for a project
func RmCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string

	rmCmd := &cobra.Command{
		Use:   "rm [<project>] <key>",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		Short: "Remove an env variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse or infer arguments
			key := args[len(args)-1]
			if len(args) == 2 {
				if projectName != "" {
					return fmt.Errorf("project name provided both as argument and flag")
				}
				projectName = args[0]
			}
			if projectName == "" {
				var err error
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Unset the variable
			client, err := ch.Client()
			if err != nil {
				return err
			}
			_, err = client.UpdateProjectVariables(cmd.Context(), &adminv1.UpdateProjectVariablesRequest{
				Org:            ch.Org,
				Project:        projectName,
				Environment:    environment,
				UnsetVariables: []string{key},
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project\n")
			return nil
		},
	}

	rmCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	rmCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	rmCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to resolve for (options: dev, prod)")

	return rmCmd
}
