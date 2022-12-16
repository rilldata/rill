package source

import (
	"fmt"
	"regexp"

	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/spf13/cobra"
)

// dropCmd represents the drop command, it requires min 1 args as source path.
func DropCmd(ver string) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var projectPath string
	var verbose bool

	dropCmd := &cobra.Command{
		Use:   "drop <source>",
		Short: "Drop a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName := args[0]

			if ok, err := regexp.Match("^[ _a-zA-Z0-9\\-]+$", []byte(sourceName)); !ok || err != nil {
				return fmt.Errorf("not a valid source name: %s", sourceName)
			}

			app, err := local.NewApp(cmd.Context(), ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			repo, err := app.Runtime.Repo(cmd.Context(), app.Instance.ID)
			if err != nil {
				panic(err) // Should never happen
			}

			c := rillv1beta.New(repo, app.Instance.ID)
			sourcePath, err := c.DeleteSource(cmd.Context(), sourceName)
			if err != nil {
				return fmt.Errorf("delete source: %w", err)
			}

			err = app.ReconcileSource(sourcePath)
			if err != nil {
				return fmt.Errorf("reconcile source: %w", err)
			}

			return nil
		},
	}

	dropCmd.Flags().SortFlags = false
	dropCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	dropCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	dropCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	dropCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return dropCmd
}
