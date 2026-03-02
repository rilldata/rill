package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

var templates = []struct {
	Name        string
	Description string
}{
	{"empty-duckdb", "Create a new empty Rill project with DuckDB"},
	{"empty-clickhouse", "Create a new empty Rill project with ClickHouse"},
	{"cursor", "Add Cursor rules to an existing Rill project"},
	{"claude", "Add Claude Code instructions to an existing Rill project"},
}

func InitCmd(ch *cmdutil.Helper) *cobra.Command {
	var template string
	var force bool

	var b strings.Builder
	b.WriteString("Initialize a new Rill project or add files to an existing project from a template.")
	b.WriteString("\n\nThe available templates are:\n")
	for _, t := range templates {
		fmt.Fprintf(&b, "- %s: %s.\n", t.Name, t.Description)
	}
	long := b.String()

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Add Rill project files from a template",
		Long:  long,
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

			// If no template specified, prompt interactively
			if template == "" {
				if !ch.Interactive {
					return fmt.Errorf("template must be specified in non-interactive mode")
				}
				names := make([]string, len(templates))
				descs := make([]string, len(templates))
				for i, t := range templates {
					names[i] = t.Name
					descs[i] = t.Description
				}
				selected, err := cmdutil.SelectPromptWithDescriptions("Select a template", names, descs, names[0])
				if err != nil {
					return err
				}
				template = selected
			}

			switch template {
			case "empty-duckdb", "empty-clickhouse":
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

				// Map template name to OLAP engine: "empty-duckdb" -> "duckdb"
				olap := strings.TrimPrefix(template, "empty-")
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
				ch.Printf("Added Claude instructions in .claude and .mcp.json\n")

			default:
				return fmt.Errorf("unknown template: %s", template)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&template, "template", "", "Project template to use (default: prompt to select)")
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files when unpacking a template")

	return initCmd
}
