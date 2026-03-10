package initialize

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/examples"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func InitCmd(ch *cmdutil.Helper) *cobra.Command {
	var olap string
	var example string
	var agent string

	exampleOptions, err := examples.List()
	if err != nil {
		ch.Printf("Warning: failed to list example projects: %v\n", err)
	}

	var long strings.Builder
	long.WriteString("Initialize a new Rill project. Use flags to customize the project or run interactively to be prompted for each option.")
	if len(exampleOptions) > 0 {
		long.WriteString("\n\nAvailable example projects:\n")
		for _, ex := range exampleOptions {
			fmt.Fprintf(&long, "  - %s (%s)\n", ex.Name, ex.OLAPConnector)
		}
	}

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Initialize a new Rill project",
		Long:  long.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Assess what flags were set
			numFlags := 0
			explicitOlap := false
			explicitAgent := false
			if cmd.Flags().Changed("olap") {
				numFlags++
				explicitOlap = true
			}
			if cmd.Flags().Changed("example") {
				numFlags++
			}
			if cmd.Flags().Changed("agent") {
				numFlags++
				explicitAgent = true
			}

			// Resolve project path:
			// - If a path arg is provided, use it directly.
			// - If cwd contains rill.yaml, default to cwd.
			// - Otherwise prompt interactively.
			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
			} else if cmdutil.HasRillProject(".") {
				projectPath = "."
			} else {
				if !ch.Interactive {
					return fmt.Errorf("project path argument is required when not running interactively")
				}
				name, err := cmdutil.InputPrompt("Project name", "my-rill-project")
				if err != nil {
					return err
				}
				projectPath = filepath.Join(".", name)
			}

			// Normalize project path
			projectPath, err := fileutil.ExpandHome(projectPath)
			if err != nil {
				return fmt.Errorf("failed to expand path %q: %w", projectPath, err)
			}
			projectPath, err = filepath.Abs(projectPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", projectPath, err)
			}

			// Infer project name
			projectName := filepath.Base(projectPath)

			// If a project already exists, we allow adding agent files via --agent, but no other changes.
			if cmdutil.HasRillProject(projectPath) {
				if explicitAgent {
					if numFlags > 1 {
						return fmt.Errorf("when adding agent instructions to an existing project, --agent must be the only flag set")
					}
					repo, _, err := cmdutil.RepoForProjectPath(projectPath)
					if err != nil {
						return fmt.Errorf("failed to open project: %w", err)
					}
					return writeAgentInstructions(cmd.Context(), ch, repo, agent)
				}
				return fmt.Errorf("a Rill project already exists at %q (use --agent to update agent instructions)", projectPath)
			}

			// In interactive mode, if no flags were provided, we prompt for input.
			// If one or more flags were provided, we don't prompt because the user
			// has already made an active choice and presumably wants defaults for
			// the remaining options.
			if ch.Interactive && numFlags == 0 {
				// OLAP
				var err error
				olap, err = cmdutil.SelectPrompt("OLAP engine", []string{"duckdb", "clickhouse"}, "duckdb")
				if err != nil {
					return err
				}

				// Example project
				examplesForOLAP := []string{"none"}
				for _, ex := range exampleOptions {
					if ex.OLAPConnector == olap {
						examplesForOLAP = append(examplesForOLAP, ex.Name)
					}
				}
				if len(examplesForOLAP) > 1 {
					selected, err := cmdutil.SelectPrompt("Example project", examplesForOLAP, "none")
					if err != nil {
						return err
					}
					if selected != "none" {
						example = selected
					}
				}

				// Agent instructions
				agent, err = cmdutil.SelectPrompt("Agent instructions", []string{"claude", "cursor", "all", "none"}, "claude")
				if err != nil {
					return err
				}

				// Print an empty line for nicer output
				ch.Printf("\n")
			}

			// Validate fields before creating any files
			if !slices.Contains(olapOptions, olap) {
				return fmt.Errorf("invalid --olap value %q (options: %s)", olap, strings.Join(olapOptions, ", "))
			}
			if !slices.Contains(agentOptions, agent) {
				return fmt.Errorf("invalid --agent value %q (options: %s)", agent, strings.Join(agentOptions, ", "))
			}
			if example != "" {
				var found bool
				for _, ex := range exampleOptions {
					if ex.Name != example {
						continue
					}
					if explicitOlap && ex.OLAPConnector != olap {
						return fmt.Errorf("example project %q is not compatible with OLAP engine %q", example, olap)
					}
					found = true
					break
				}
				if !found {
					return fmt.Errorf("invalid --example value %q (see help menu for options)", example)
				}
			}

			// Create project directory
			if err := os.MkdirAll(projectPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", projectPath, err)
			}

			// Open project repo
			repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
			if err != nil {
				return fmt.Errorf("failed to open project: %w", err)
			}

			// Initialize empty project
			if err := parser.InitEmpty(cmd.Context(), repo, instanceID, projectName, olap); err != nil {
				return fmt.Errorf("failed to create empty project: %w", err)
			}
			ch.Printf("Created a new Rill project at %s\n", projectPath)

			// Unpack example files
			if example != "" {
				exampleFS, err := examples.Get(example)
				if err != nil {
					return fmt.Errorf("failed to get example project %q: %w", example, err)
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
					return repo.Put(cmd.Context(), p, file)
				})
				if err != nil {
					return fmt.Errorf("failed to unpack example project: %w", err)
				}
				ch.Printf("Unpacked example project %q\n", example)
			}

			// Write agent files
			if err := writeAgentInstructions(cmd.Context(), ch, repo, agent); err != nil {
				return err
			}

			// Print next steps
			projectPathRelative := projectPath
			cwd, err := os.Getwd()
			if err != nil {
				cwd = "." // Safe but usually ineffective fallback
			}
			if rel, err := filepath.Rel(cwd, projectPath); err == nil {
				projectPathRelative = rel
			}
			escaped := fileutil.ShellEscape(projectPathRelative)
			if ch.Interactive {
				ch.Printf("\nSuccess! Run the following command to start the project:\n\n")
				ch.Printf("  rill start %s\n\n", escaped)
			} else {
				ch.Printf("Run `rill validate %s` to build and validate the project, or `rill start %s` to build and serve the project on localhost\n", escaped, escaped)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&olap, "olap", "duckdb", fmt.Sprintf("OLAP engine (options: %s)", strings.Join(olapOptions, ", ")))
	initCmd.Flags().StringVar(&example, "example", "", "Example project name (default: empty project)")
	initCmd.Flags().StringVar(&agent, "agent", "claude", fmt.Sprintf("Agent instructions (options: %s)", strings.Join(agentOptions, ", ")))

	return initCmd
}

// olapOptions lists the supported OLAP engines.
var olapOptions = []string{
	"duckdb",
	"clickhouse",
}

// agentOptions lists the supported agent instruction sets.
var agentOptions = []string{
	"claude",
	"cursor",
	"all",
	"none",
}

// writeAgentInstructions initializes agent instruction files based on the selected agent type.
func writeAgentInstructions(ctx context.Context, ch *cmdutil.Helper, repo drivers.RepoStore, agent string) error {
	switch agent {
	case "all":
		if err := instructions.InitClaudeCode(ctx, repo, true); err != nil {
			return fmt.Errorf("failed to add Claude Code files: %w", err)
		}
		ch.Printf("Added Claude instructions in .claude and .mcp.json\n")
		if err := instructions.InitCursorRules(ctx, repo, true); err != nil {
			return fmt.Errorf("failed to add Cursor rules: %w", err)
		}
		ch.Printf("Added Cursor rules in .cursor\n")
	case "claude":
		if err := instructions.InitClaudeCode(ctx, repo, true); err != nil {
			return fmt.Errorf("failed to add Claude Code files: %w", err)
		}
		ch.Printf("Added Claude instructions in .claude and .mcp.json\n")
	case "cursor":
		if err := instructions.InitCursorRules(ctx, repo, true); err != nil {
			return fmt.Errorf("failed to add Cursor rules: %w", err)
		}
		ch.Printf("Added Cursor rules in .cursor\n")
	case "none":
		// No agent instructions to add
	default:
		return fmt.Errorf("invalid agent option %q", agent)
	}
	return nil
}
