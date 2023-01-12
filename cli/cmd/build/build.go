package build

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func BuildCmd(ver version.Version) *cobra.Command {
	var projectPath string
	var olapDriver string
	var olapDSN string
	var verbose bool
	var logFormat string

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build project without starting web app",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := local.NewApp(cmd.Context(), ver, verbose, olapDriver, olapDSN, projectPath, logFormat)
			if err != nil {
				return err
			}
			defer app.Close()

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			err = app.Reconcile()
			if err != nil {
				return fmt.Errorf("reconcile project: %w", err)
			}

			return nil
		},
	}
	buildCmd.Flags().SortFlags = false
	buildCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	buildCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	buildCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	buildCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	buildCmd.Flags().StringVar(&logFormat, "log-format", "color", "Sets the log format to color/json")

	return buildCmd
}
