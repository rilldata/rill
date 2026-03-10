package initialize

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/examples"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func InitCmd(ch *cmdutil.Helper) *cobra.Command {
	var olap string
	var demo string
	var agent string
	var nonInteractive bool
	var force bool

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Initialize a new Rill project",
		Long:  "Initialize a new Rill project. Use flags to customize the project or run interactively to be prompted for each option.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			interactive := ch.Interactive && !nonInteractive

			// Validate flag values
			if cmd.Flags().Changed("olap") {
				if olap != "duckdb" && olap != "clickhouse" {
					return fmt.Errorf("invalid --olap value %q: must be \"duckdb\" or \"clickhouse\"", olap)
				}
			}
			if cmd.Flags().Changed("agent") {
				if agent != "claude" && agent != "cursor" && agent != "all" && agent != "none" {
					return fmt.Errorf("invalid --agent value %q: must be \"claude\", \"cursor\", \"all\", or \"none\"", agent)
				}
			}
			if cmd.Flags().Changed("demo") && cmd.Flags().Changed("olap") && olap != "duckdb" {
				return fmt.Errorf("--demo is only supported with --olap duckdb")
			}

			// Resolve project path
			var projectName string
			var projectPath string
			if len(args) > 0 {
				projectName = args[0]
			} else if interactive {
				var err error
				projectName, err = cmdutil.InputPrompt("Project name", "my-rill-project")
				if err != nil {
					return err
				}
			} else {
				projectName = "my-rill-project"
			}

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

			// When args are provided, targetPath IS the project path.
			// When no args, the project is created as a subdirectory.
			if len(args) > 0 {
				projectPath = targetPath
			} else {
				projectPath = filepath.Join(targetPath, projectName)
			}

			// Resolve OLAP engine
			if !cmd.Flags().Changed("olap") {
				if interactive {
					olap, err = cmdutil.SelectPrompt("OLAP engine", []string{"duckdb", "clickhouse"}, "duckdb")
					if err != nil {
						return err
					}
				}
				// else: use default "duckdb"
			}

			// Resolve demo project (DuckDB only)
			if !cmd.Flags().Changed("demo") && olap == "duckdb" {
				if interactive {
					demoList, err := examples.List()
					if err != nil {
						return fmt.Errorf("failed to list demo projects: %w", err)
					}
					options := []string{"None"}
					for _, ex := range demoList {
						options = append(options, ex.Name)
					}
					selected, err := cmdutil.SelectPrompt("Use demo project?", options, "None")
					if err != nil {
						return err
					}
					if selected != "None" {
						demo = selected
					}
				}
				// else: use default "" (no demo)
			}

			// Validate demo against olap (for the case where olap was resolved via prompt)
			if demo != "" && olap != "duckdb" {
				return fmt.Errorf("--demo is only supported with --olap duckdb")
			}

			// Validate demo name if specified
			if demo != "" {
				demoList, err := examples.List()
				if err != nil {
					return fmt.Errorf("failed to list demo projects: %w", err)
				}
				found := false
				for _, ex := range demoList {
					if ex.Name == demo {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("unknown demo project %q", demo)
				}
			}

			// Resolve agent
			if !cmd.Flags().Changed("agent") {
				if interactive {
					agent, err = cmdutil.SelectPrompt("Agent", []string{"claude", "cursor", "all", "none"}, "claude")
					if err != nil {
						return err
					}
				}
				// else: use default "claude"
			}

			// Create project directory
			if cmdutil.HasRillProject(projectPath) && !force {
				return fmt.Errorf("a Rill project already exists at %q", projectPath)
			}
			if err := os.MkdirAll(projectPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", projectPath, err)
			}

			repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
			if err != nil {
				return fmt.Errorf("failed to initialize repo: %w", err)
			}

			// Initialize empty project
			err = parser.InitEmpty(ctx, repo, instanceID, projectName, olap)
			if err != nil {
				return fmt.Errorf("failed to create empty project: %w", err)
			}
			ch.Printf("Created a new Rill project at %q\n", projectPath)

			// Unpack demo files if selected
			if demo != "" {
				exampleFS, err := examples.Get(demo)
				if err != nil {
					return fmt.Errorf("failed to get demo project %q: %w", demo, err)
				}
				err = fs.WalkDir(exampleFS, ".", func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					file, err := exampleFS.Open(p)
					if err != nil {
						return err
					}
					defer file.Close()
					return repo.Put(ctx, p, file)
				})
				if err != nil {
					return fmt.Errorf("failed to unpack demo project: %w", err)
				}
				ch.Printf("Unpacked demo project %q\n", demo)
			}

			// Initialize agent files
			switch agent {
			case "claude":
				err = instructions.InitClaudeCode(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Claude Code files: %w", err)
				}
				ch.Printf("Added Claude instructions in .claude and .mcp.json\n")
			case "cursor":
				err = instructions.InitCursorRules(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Cursor rules: %w", err)
				}
				ch.Printf("Added Cursor rules in .cursor\n")
			case "all":
				err = instructions.InitClaudeCode(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Claude Code files: %w", err)
				}
				ch.Printf("Added Claude instructions in .claude and .mcp.json\n")
				err = instructions.InitCursorRules(ctx, repo, force)
				if err != nil {
					return fmt.Errorf("failed to add Cursor rules: %w", err)
				}
				ch.Printf("Added Cursor rules in .cursor\n")
			}

			// In non-interactive mode, we're done
			if !interactive {
				return nil
			}

			// Prompt: Start Rill?
			startRill, err := cmdutil.ConfirmPrompt("Start Rill?", "", true)
			if err != nil {
				return err
			}
			if startRill {
				startCmd, _, err := cmd.Root().Find([]string{"start"})
				if err != nil {
					return fmt.Errorf("failed to find start command: %w", err)
				}
				startCmd.SetContext(ctx)
				return startCmd.RunE(startCmd, []string{projectPath})
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&olap, "olap", "duckdb", "OLAP engine: \"duckdb\" or \"clickhouse\"")
	initCmd.Flags().StringVar(&demo, "demo", "", "Demo project name (DuckDB only); empty means none")
	initCmd.Flags().StringVar(&agent, "agent", "claude", "Agent instructions: \"claude\", \"cursor\", \"all\", or \"none\"")
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Use defaults for unspecified flags; do not start Rill")
	initCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")

	return initCmd
}
