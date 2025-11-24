package initialize

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/spf13/cobra"
)

//go:embed templates/**/*.mdc
var templateFS embed.FS

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

	// Walk the embedded templates/cursor directory and copy files into .cursor/rules
	return fs.WalkDir(templateFS, "templates/cursor", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := filepath.Base(path)                             // e.g. "code-style.md"
		relativePath := filepath.Join(".cursor", "rules", name) // target relative path
		fp := filepath.Join(basePath, relativePath)

		contentB, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := os.Stat(fp); err == nil && !force {
			return nil
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}

		return os.WriteFile(fp, contentB, 0o644)
	})
}
