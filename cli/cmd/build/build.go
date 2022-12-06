package build

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

func BuildCmd(ver string) *cobra.Command {
	var projectPath string
	var olapDriver string
	var olapDSN string
	var verbose bool

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build project without starting web app",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := local.NewApp(ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			err = app.Reconcile()
			if err != nil {
				return err
			}

			return nil
		},
	}
	buildCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "OLAP database driver")
	buildCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "OLAP database DSN")
	buildCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	buildCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return buildCmd
}
