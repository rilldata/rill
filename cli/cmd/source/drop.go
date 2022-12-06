package source

import (
	"context"
	"fmt"
	"regexp"

	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/artifacts/artifactsv0"
	"github.com/spf13/cobra"
)

// dropCmd represents the drop command, it requires min 1 args as source path
func DropCmd(ver string) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var projectPath string
	var verbose bool

	var dropCmd = &cobra.Command{
		Use:   "drop <source>",
		Short: "Drop a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName := args[0]

			if ok, err := regexp.Match("^[ _a-zA-Z0-9\\-]+$", []byte(sourceName)); !ok || err != nil {
				return fmt.Errorf("not a valid source name: %s", sourceName)
			}

			app, err := local.NewApp(ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			repo, err := app.Runtime.Repo(context.Background(), app.Instance.ID)
			if err != nil {
				panic(err) // Should never happen
			}

			c := artifactsv0.New(repo, app.Instance.ID)
			sourcePath, err := c.DeleteSource(context.Background(), sourceName)
			if err != nil {
				return err
			}

			err = app.ReconcileSource(sourcePath)
			if err != nil {
				return err
			}

			return nil
		},
	}

	dropCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "OLAP database driver")
	dropCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "OLAP database DSN")
	dropCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	dropCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return dropCmd
}
