package env

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/spf13/cobra"
)

func PushCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string

	pushCmd := &cobra.Command{
		Use:   "push [<project-name>]",
		Short: "Push local .env contents to cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectPath != "" {
				var err error
				projectPath, err = normalizeProjectPath(projectPath)
				if err != nil {
					return fmt.Errorf("failed to normalize project path: %w", err)
				}
			}

			if len(args) > 0 {
				projectName = args[0]
			}

			// Parse and verify the project directory
			repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
			if err != nil {
				return fmt.Errorf("failed to get repo for project path: %w", err)
			}
			p, err := parser.Parse(cmd.Context(), repo, instanceID, "prod", "duckdb")
			if err != nil {
				return fmt.Errorf("failed to parse project: %w", err)
			}
			if p.RillYAML == nil {
				return fmt.Errorf("not a valid Rill project (missing a rill.yaml file)")
			}

			// Find the cloud project name
			if projectName == "" {
				projectName, err = ch.InferProjectName(cmd.Context(), ch.Org, projectPath)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Fetch the project variables from the cloud
			client, err := ch.Client()
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			res, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				Org:         ch.Org,
				Project:     projectName,
				Environment: environment,
			})
			if err != nil {
				return fmt.Errorf("failed to get project variables: %w", err)
			}

			// new vars from the cloud
			perEnvVars := make(map[string]map[string]string)
			for _, v := range res.Variables {
				vars, ok := perEnvVars[v.Environment]
				if !ok {
					vars = make(map[string]string)
					perEnvVars[v.Environment] = vars
				}
				vars[v.Name] = v.Value
			}

			// existing vars from the .env files in the project
			current := p.GetDotEnvPerEnvironment()

			// Merge the current .env file with the cloud variables
			for env, local := range current {
				if env != "" && env != environment {
					ch.Printf("Skipping environment %q since it doesn't match the specified environment filter %q.\n", env, environment)
					continue
				}
				cloud, ok := perEnvVars[env]
				if !ok {
					cloud = make(map[string]string)
				}
				var added, changed int
				for k, v := range local {
					if _, ok := cloud[k]; !ok {
						added++
					} else if cloud[k] != v {
						changed++
					}
					cloud[k] = v
				}
				// no changes
				if added+changed == 0 {
					ch.Printf("Environment %q: There are no new or changed variables in your local file.\n", envForPrint(env))
					continue
				}

				if added > 0 || changed > 0 {
					ch.Printf("Environment %q: %d new and %d changed variable(s) found in local file.\n", envForPrint(env), added, changed)
				}
				confirmed := true
				if ch.Interactive {
					confirmed, err = cmdutil.ConfirmPrompt("Do you want to continue?", "", true)
					if err != nil {
						return fmt.Errorf("failed to prompt for confirmation: %w", err)
					}
				}
				if !confirmed {
					continue
				}

				// Write the merged variables back to the cloud project
				_, err = client.UpdateProjectVariables(cmd.Context(), &adminv1.UpdateProjectVariablesRequest{
					Org:         ch.Org,
					Project:     projectName,
					Environment: env,
					Variables:   cloud,
				})
				if err != nil {
					return fmt.Errorf("failed to update project variables: %w", err)
				}

				ch.Printf("Environment %q: Updated cloud env for project %q with variables from %q.\n", env, projectName, pathForEnv(env))
			}
			return nil
		},
	}

	pushCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	pushCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	pushCmd.Flags().StringVar(&environment, "environment", "dev", "Optional environment to resolve for (options: dev, prod)")

	return pushCmd
}

func envForPrint(env string) string {
	if env == "" {
		return "default"
	}
	return env
}
