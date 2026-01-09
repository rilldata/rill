package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func InitCmd(ch *cmdutil.Helper) *cobra.Command {
	var template string
	var force bool

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Add Rill project files from a template",
		Long: `Initialize a new Rill project or add files to an existing project from a template.

The available templates are:
- duckdb: Creates an empty Rill project configured to use DuckDB as the OLAP database.
- clickhouse: Creates an empty Rill project configured to use ClickHouse as the OLAP database.
- cursor: Adds Cursor rules in .cursor to an existing Rill project.
- claude: Adds Claude Code instruction in .claude to an existing Rill project.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			targetPath := "."
			if len(args) > 0 {
				targetPath = args[0]
			}
			targetPath, err := fileutil.ExpandHome(targetPath)
			if err != nil {
				return fmt.Errorf("failed to expand path %q: %w", targetPath, err)
			}
			targetPath, err = filepath.Abs(targetPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", targetPath, err)
			}

			switch template {
			case "duckdb", "clickhouse":
				if cmdutil.HasRillProject(targetPath) {
					return fmt.Errorf("a Rill project already exists at %q", targetPath)
				}

				if err := os.MkdirAll(targetPath, 0o755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
				}

				repo, instanceID, err := cmdutil.RepoForProjectPath(targetPath)
				if err != nil {
					return fmt.Errorf("failed to initialize repo: %w", err)
				}

				olap := template // Currently map 1:1 with template
				err = parser.InitEmpty(ctx, repo, instanceID, "My Rill project", olap)
				if err != nil {
					return fmt.Errorf("failed to create empty project: %w", err)
				}
				ch.Printf("Created a new Rill project at %q\n", targetPath)

			case "cursor":
				if !cmdutil.HasRillProject(targetPath) {
					return fmt.Errorf("no Rill project found at %q; run `rill init` first to create an empty project.", targetPath)
				}
				repo, _, err := cmdutil.RepoForProjectPath(targetPath)
				if err != nil {
					return fmt.Errorf("failed to initialize repo: %w", err)
				}

				err = instructions.InitCursorRules(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Cursor rules: %w", err)
				}
				ch.Printf("Added Cursor rules in .cursor\n")

			case "claude":
				if !cmdutil.HasRillProject(targetPath) {
					return fmt.Errorf("no Rill project found at %q; run `rill init` first to create an empty project.", targetPath)
				}
				repo, _, err := cmdutil.RepoForProjectPath(targetPath)
				if err != nil {
					return fmt.Errorf("failed to initialize repo: %w", err)
				}

				err = instructions.InitClaudeCode(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Claude Code files: %w", err)
				}
				ch.Printf("Added Claude instructions in .claude\n")

			default:
				return fmt.Errorf("unknown template: %s", template)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&template, "template", "duckdb", "Project template to use (options: duckdb, clickhouse, cursor)")
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files when unpacking a template")

	return initCmd
}
