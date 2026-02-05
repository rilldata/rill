package env

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/spf13/cobra"
)

func PullCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, environment string

	pullCmd := &cobra.Command{
		Use:   "pull [<project-name>]",
		Short: "Pull cloud credentials into local .env file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If projectPath is provided, normalize it
			if projectPath != "" {
				var err error
				projectPath, err = normalizeProjectPath(projectPath)
				if err != nil {
					return err
				}
			}

			if len(args) > 0 {
				projectName = args[0]
			}

			// Fetch the project variables from the cloud
			return PullVars(cmd.Context(), ch, projectPath, projectName, environment, true)
		},
	}

	pullCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	pullCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")
	pullCmd.Flags().StringVar(&environment, "environment", "dev", "Environment to resolve for (options: dev, prod)")

	return pullCmd
}

func PullVars(ctx context.Context, ch *cmdutil.Helper, projectPath, projectName, environment string, warnForNoVars bool) error {
	// Parse and verify the project directory
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return err
	}
	p, err := parser.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return fmt.Errorf("failed to parse project: %w", err)
	}
	if p.RillYAML == nil {
		return fmt.Errorf("not a valid Rill project (missing a rill.yaml file)")
	}

	// Find the cloud project name
	if projectName == "" {
		projectName, err = ch.InferProjectName(ctx, ch.Org, projectPath)
		if err != nil {
			return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
		}
	}
	client, err := ch.Client()
	if err != nil {
		return err
	}
	res, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
		Org:         ch.Org,
		Project:     projectName,
		Environment: environment,
	})
	if err != nil {
		return err
	}

	resVars := make(map[string]string, len(res.Variables))
	for _, v := range res.Variables {
		resVars[v.Name] = v.Value
	}

	dotEnv := p.GetDotEnv()

	// If the variables match any existing .env file, do nothing
	if maps.Equal(resVars, dotEnv) && warnForNoVars {
		if len(res.Variables) == 0 {
			ch.Printf("No cloud credentials found for project %q.\n", projectName)
		} else {
			ch.Printf("Local .env file is already up to date with cloud credentials.\n")
		}
		return nil
	}

	// Merge the current .env file with pulled variables
	vars := make(map[string]string)
	maps.Copy(vars, dotEnv)
	maps.Copy(vars, resVars)
	err = godotenv.Write(vars, filepath.Join(projectPath, ".env"))
	if err != nil {
		return err
	}

	// Add to gitignore if necessary
	changed, err := cmdutil.EnsureGitignoreHasDotenv(ctx, repo)
	if err != nil {
		return err
	}
	if changed {
		ch.Printf("Added .env to .gitignore.\n")
	}

	ch.Printf("Updated .env file with cloud credentials from project %q.\n", projectName)
	return nil
}
