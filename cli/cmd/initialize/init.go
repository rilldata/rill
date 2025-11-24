package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/parser"
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
				projectPath := filepath.Clean(args[0])
				absPath, err := filepath.Abs(projectPath)
				if err != nil {
					return fmt.Errorf("failed to resolve path %q: %w", projectPath, err)
				}
				if err := os.MkdirAll(absPath, 0o755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", absPath, err)
				}
				targetPath = absPath
			}
			isInProject := cmdutil.HasRillProject(targetPath)

			switch template {
			case "default":
				if isInProject {
					ch.Printf("Already in a Rill project\n")
					return nil
				}
				err := parser.InitEmpty(ctx, nil, "", "Rill project", olap)
				if err != nil {
					return fmt.Errorf("failed to create empty project: %w", err)
				}
				ch.Printf("Created empty Rill project\n")

			case "cursor":
				if !isInProject {
					err := parser.InitEmpty(ctx, nil, "", "Rill project", olap)
					if err != nil {
						return fmt.Errorf("failed to create empty project: %w", err)
					}
				}

				if force && ch.Interactive {
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

				err := addCursorRules(force, targetPath)
				if err != nil {
					return fmt.Errorf("failed to add Cursor rules: %w", err)
				}
			default:
				return fmt.Errorf("unknown template: %s", template)
			}

			return nil
		},
	}

	initCmd.PersistentFlags().StringVar(&template, "template", "default", "Project template to use (default|cursor)")
	initCmd.PersistentFlags().StringVar(&olap, "olap", "duckdb", "OLAP engine to use (duckdb|clickhouse)")
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files when adding templates/rules")

	return initCmd
}

func addCursorRules(force bool, basePath string) error {
	if basePath == "" {
		basePath = "."
	}
	// Create .cursor/rules directory
	if err := os.MkdirAll(filepath.Join(basePath, ".cursor", "rules"), 0o755); err != nil {
		return err
	}

	files := map[string]string{
		".cursor/rules/code-style.md": `# Code Style
- Use SQL for defining models in .sql files
- Use YAML for configuration files like rill.yaml, sources, dashboards
- Follow consistent naming conventions: snake_case for SQL, camelCase for YAML keys
`,
		".cursor/rules/project-structure.md": `# Project Structure
- Place models in models/ directory
- Place dashboards in dashboards/ directory
- Place sources in sources/ directory
- Keep rill.yaml in the root
`,
		".cursor/rules/best-practices.md": `# Best Practices
- Always define sources before models
- Use metrics views for aggregations
- Test your dashboards after changes
`,
	}

	for path, content := range files {
		fp := filepath.Join(basePath, path)
		if _, err := os.Stat(fp); err == nil {
			// file exists
			if !force {
				// skip existing files
				continue
			}
			if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
				return err
			}
			continue
		} else if !os.IsNotExist(err) {
			return err
		}

		// file doesn't exist so create
		if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}
