package python

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/drivers/python"
	"github.com/spf13/cobra"
)

// packageSets defines curated package bundles available during setup.
var packageSets = []struct {
	Name        string
	Description string
	Packages    []string
}{
	{"stripe", "Stripe API (billing, charges, subscriptions)", []string{"stripe"}},
	{"google-analytics", "Google Analytics Data API (GA4)", []string{"google-analytics-data"}},
	{"dbt", "dbt Core with DuckDB adapter", []string{"dbt-core", "dbt-duckdb"}},
	{"requests", "HTTP requests library (general REST APIs)", []string{"requests"}},
}

func SetupCmd(ch *cmdutil.Helper) *cobra.Command {
	var pythonPath string
	var packages string
	var noVenv bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up a Python environment for data sources",
		Long:  "Detect or install Python, create a virtual environment, and install packages for use with Rill Python sources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve project directory
			projectRoot, err := os.Getwd()
			if err != nil {
				return err
			}

			if !cmdutil.HasRillProject(projectRoot) {
				return fmt.Errorf("no Rill project found in %s (run 'rill init' first)", projectRoot)
			}

			// Step 1: Detect Python
			fmt.Println("Detecting Python installation...")
			info, err := python.DetectPython(pythonPath)
			if err != nil {
				fmt.Printf("  Python not found: %s\n", err)
				fmt.Println("\n  Install Python 3.9+ and try again.")
				fmt.Println("  Recommended: https://github.com/pyenv/pyenv#installation")
				return fmt.Errorf("python not found")
			}
			fmt.Printf("  Found Python %s at %s\n\n", info.Version, info.Path)

			if noVenv {
				// Write connector YAML without venv
				return writeConnectorYAML(projectRoot, info.Path)
			}

			// Step 2: Select packages
			var selectedPackages []string
			if packages != "" {
				// Non-interactive: use flag value
				selectedPackages = strings.Split(packages, ",")
				for i := range selectedPackages {
					selectedPackages[i] = strings.TrimSpace(selectedPackages[i])
				}
			} else {
				// Interactive: prompt for each package set
				fmt.Println("Select additional packages to install (pandas + pyarrow are always included):")
				for _, ps := range packageSets {
					install, err := cmdutil.ConfirmPrompt(
						fmt.Sprintf("  Install %s (%s)?", ps.Name, ps.Description),
						"",
						false,
					)
					if err != nil {
						return err
					}
					if install {
						selectedPackages = append(selectedPackages, ps.Packages...)
					}
				}

				// Offer custom packages
				custom, err := cmdutil.InputPrompt("Additional pip packages (comma-separated, or empty to skip)", "")
				if err != nil {
					return err
				}
				if custom != "" {
					for _, pkg := range strings.Split(custom, ",") {
						pkg = strings.TrimSpace(pkg)
						if pkg != "" {
							selectedPackages = append(selectedPackages, pkg)
						}
					}
				}
			}

			// Step 3: Create venv and install packages
			fmt.Println("\nSetting up Python environment...")
			result, err := python.SetupEnvironment(cmd.Context(), &python.SetupOptions{
				ProjectRoot: projectRoot,
				Packages:    selectedPackages,
				PythonPath:  info.Path,
			})
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			fmt.Printf("  Virtual environment: %s\n", result.VenvPath)
			fmt.Printf("  Python: %s\n", result.PythonPath)
			fmt.Printf("  Installed: %s\n\n", strings.Join(result.Installed, ", "))

			// Step 4: Write connector YAML
			err = writeConnectorYAML(projectRoot, result.PythonPath)
			if err != nil {
				return err
			}

			fmt.Println("Python environment is ready!")
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Create a script in scripts/ that writes data to RILL_OUTPUT_PATH")
			fmt.Println("  2. Create a model YAML with 'connector: python' and 'code_path: scripts/your_script.py'")
			fmt.Println("  3. Run 'rill start' to see your data")

			return nil
		},
	}

	cmd.Flags().StringVar(&pythonPath, "python-path", "", "Path to Python executable (default: auto-detect)")
	cmd.Flags().StringVar(&packages, "packages", "", "Comma-separated pip packages to install (non-interactive)")
	cmd.Flags().BoolVar(&noVenv, "no-venv", false, "Skip virtual environment creation; use system Python directly")

	return cmd
}

func writeConnectorYAML(projectRoot, pythonPath string) error {
	connectorDir := filepath.Join(projectRoot, "connectors")
	if err := os.MkdirAll(connectorDir, 0o755); err != nil {
		return fmt.Errorf("failed to create connectors directory: %w", err)
	}

	connectorPath := filepath.Join(connectorDir, "python.yaml")
	content := fmt.Sprintf("type: connector\ndriver: python\npython_path: %s\n", pythonPath)
	err := os.WriteFile(connectorPath, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write connector YAML: %w", err)
	}
	fmt.Printf("  Wrote %s\n", connectorPath)
	return nil
}
