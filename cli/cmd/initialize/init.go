package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func InitCmd(ch *cmdutil.Helper) *cobra.Command {
	var olap, template string
	var force bool

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Initialize Rill resources from templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			targetPath := "."

			if len(args) > 0 {
				projectPath, err := fileutil.ExpandHome(args[0])
				if err != nil {
					return fmt.Errorf("failed to expand path %q: %w", args[0], err)
				}
				absPath, err := filepath.Abs(projectPath)
				if err != nil {
					return fmt.Errorf("failed to resolve path %q: %w", projectPath, err)
				}
				targetPath = absPath
			}
			isInProject := cmdutil.HasRillProject(targetPath)

			if template == "cursor" && force && ch.Interactive {
				ch.PrintfWarn("Warning: --force will overwrite existing Cursor rule files.\n")
				ok, err := cmdutil.ConfirmPrompt("This will overwrite existing rule files. Continue?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			switch template {
			case "default":
				if isInProject {
					ch.Printf("Already in a Rill project\n")
					return nil
				}

				if err := os.MkdirAll(targetPath, 0o755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
				}

				repo, instanceID, err := cmdutil.RepoForProjectPath(targetPath)
				if err != nil {
					return fmt.Errorf("failed to initialize repo: %w", err)
				}

				err = parser.InitEmpty(ctx, repo, instanceID, "Rill project", olap)
				if err != nil {
					return fmt.Errorf("failed to create empty project: %w", err)
				}
				ch.Printf("Created empty Rill project\n")

			case "cursor":
				if !isInProject {
					return fmt.Errorf("cannot add Cursor rules: not in a Rill project. Run 'rill init' first to create a project")
				}

				repo, _, err := cmdutil.RepoForProjectPath(targetPath)
				if err != nil {
					return fmt.Errorf("failed to initialize repo: %w", err)
				}

				err = parser.InitCursorRules(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Cursor rules: %w", err)
				}
				ch.Printf("Added Cursor rules to .cursor/rules/\n")
			default:
				return fmt.Errorf("unknown template: %s", template)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&template, "template", "default", "Project template to use (default|cursor)")
	initCmd.Flags().StringVar(&olap, "olap", "duckdb", "OLAP engine to use (duckdb|clickhouse)")
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files when adding templates/rules")

	return initCmd
}
