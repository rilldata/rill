package build

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

func BuildCmd(cfg *config.Config) *cobra.Command {
	var projectPath string
	var olapDriver string
	var olapDSN string
	var verbose bool
	var variables []string

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build project without starting web app",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := local.NewApp(cmd.Context(), cfg.Version, verbose, olapDriver, olapDSN, projectPath, local.LogFormatConsole, variables)
			if err != nil {
				return err
			}
			defer app.Close()

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			err = app.Reconcile(true)
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
	buildCmd.Flags().StringSliceVarP(&variables, "env", "e", []string{}, "Set project variables")

	return buildCmd
}
