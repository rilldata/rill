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
	var projectPath, projectName, environment, keyValPair string

	setCmd := &cobra.Command{
		Use:   "set [<project-name>] [--env=key=value]",
		Args:  cobra.ExactArgs(1),
		Short: "Set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				projectName = args[0]
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			key, value, err := envValidator.ParseAndValidate(keyValPair)
			if err != nil {
				return err
			}

			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Organization: ch.Org,
				Project:      projectName,
				Environment:  environment,
				Variables:    map[string]string{key: value},
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
	setCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to resolve for (options: dev, prod)")
	setCmd.Flags().StringVar(&keyValPair, "env", "", "Specify a key and value to insert in environment (i.e. somekey=somevalue)")

	return setCmd
}
