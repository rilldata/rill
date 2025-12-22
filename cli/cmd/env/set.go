package env

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	envValidator "github.com/rilldata/rill/runtime/pkg/env"
	"github.com/spf13/cobra"
)

// SetCmd is sub command for env. Sets the variable for a project
func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string
	var variables map[string]string

	setCmd := &cobra.Command{
		Use:   "set [<project>] <key> <value>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 2 {
				variables = map[string]string{args[0]: args[1]}
			} else {
				projectName, variables = args[0], map[string]string{args[1]: args[2]}
			}

			if projectName == "" && !cmd.Flags().Changed("project") {
				projectName, err = ch.InferProjectName(ctx, ch.Org, projectPath)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			err = envValidator.ValidateVariables(variables)
			if err != nil {
				return err
			}

			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Org:         ch.Org,
				Project:     projectName,
				Environment: environment,
				Variables:   variables,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project variables\n")
			return nil
		},
	}

	setCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	setCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	setCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to set for (options: dev, prod)")
	return setCmd
}
