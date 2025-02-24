package env

import (
	"fmt"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
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
			parser, err := rillv1.Parse(cmd.Context(), repo, instanceID, "prod", "duckdb")
			if err != nil {
				return fmt.Errorf("failed to parse project: %w", err)
			}
			if parser.RillYAML == nil {
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
				Organization: ch.Org,
				Project:      projectName,
				Environment:  environment,
			})
			if err != nil {
				return fmt.Errorf("failed to get project variables: %w", err)
			}

			// Merge the current .env file with the cloud variables
			vars := make(map[string]string)
			for _, v := range res.Variables {
				vars[v.Name] = v.Value
			}
			added := 0
			changed := 0
			for k, v := range parser.DotEnv {
				if _, ok := vars[k]; !ok {
					added++
				} else if vars[k] != v {
					changed++
				}
				vars[k] = v
			}

			// If there were no changes, exit early
			if added+changed == 0 {
				ch.Print("There are no new or changed variables in your local .env file.\n")
				return nil
			}

			// Always prompt for confirmation when there are changes
			message := fmt.Sprintf("Found %d new and %d changed variable(s) to push to project %q:\n", added, changed, projectName)
			ch.Print(message)

			// Add details about the changes
			for k, v := range parser.DotEnv {
				if _, ok := vars[k]; !ok {
					ch.Printf("  + %s: will be added\n", k)
				} else if vars[k] != v {
					ch.Printf("  ~ %s: will be updated\n", k)
				}
			}

			ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", true)
			if err != nil {
				return fmt.Errorf("failed to prompt for confirmation: %w", err)
			}
			if !ok {
				return nil
			}

			// Write the merged variables back to the cloud project
			_, err = client.UpdateProjectVariables(cmd.Context(), &adminv1.UpdateProjectVariablesRequest{
				Organization: ch.Org,
				Project:      projectName,
				Environment:  environment,
				Variables:    vars,
			})
			if err != nil {
				return fmt.Errorf("failed to update project variables: %w", err)
			}

			ch.Printf("Updated cloud env for project %q with variables from %q.\n", projectName, filepath.Join(projectPath, ".env"))
			return nil
		},
	}

	pushCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	pushCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	pushCmd.Flags().StringVar(&environment, "environment", "", "Optional environment to resolve for (options: dev, prod)")

	return pushCmd
}
