package env

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
)

func PullCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName string

	pullCmd := &cobra.Command{
		Use:   "pull",
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

			// Parse and verify the project directory
			repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
			if err != nil {
				return err
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
					return err
				}
			}

			// Fetch the project variables from the cloud
			client, err := ch.Client()
			if err != nil {
				return err
			}
			res, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				OrganizationName: ch.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			// If the variables match any existing .env file, do nothing
			if maps.Equal(res.Variables, parser.DotEnv) {
				if len(res.Variables) == 0 {
					ch.Printf("No cloud credentials found for project %q.\n", projectName)
				} else {
					ch.Printf("Local .env file is already up to date with cloud credentials.\n")
				}
				return nil
			}

			// Merge the current .env file with pulled variables
			vars := make(map[string]string)
			for k, v := range parser.DotEnv {
				vars[k] = v
			}
			for k, v := range res.Variables {
				vars[k] = v
			}
			err = godotenv.Write(vars, filepath.Join(projectPath, ".env"))
			if err != nil {
				return err
			}

			// Add to gitignore if necessary
			changed, err := ensureGitignoreHas(cmd.Context(), repo, ".env")
			if err != nil {
				return err
			}
			if changed {
				ch.Printf("Added .env to .gitignore.\n")
			}

			ch.Printf("Updated .env file with cloud credentials from project %q.\n", projectName)
			return nil
		},
	}

	pullCmd.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	pullCmd.Flags().StringVar(&projectName, "project", "", "Cloud project name (will attempt to infer from Git remote if not provided)")

	return pullCmd
}

var gitignoreHasDotenvRegexp = regexp.MustCompile(`(?m)^\.env$`)

func ensureGitignoreHas(ctx context.Context, repo drivers.RepoStore, line string) (bool, error) {
	// Read .gitignore
	gitignore, _ := repo.Get(ctx, ".gitignore")

	// If .gitignore already has .env, do nothing
	if gitignoreHasDotenvRegexp.MatchString(gitignore) {
		return false, nil
	}

	// Add .env to the end of .gitignore
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += line + "\n"

	// Write .gitignore
	err := repo.Put(ctx, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return false, err
	}

	return true, nil
}
