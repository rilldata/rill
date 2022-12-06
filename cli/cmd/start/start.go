package start

import (
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// StartCmd represents the start command
func StartCmd(ver string) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var projectPath string
	var httpPort int
	var grpcPort int
	var verbose bool
	var noUI bool
	var noOpen bool

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Build project and start web application",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := local.NewApp(ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			// If not initialized, init repo with an empty project
			if !app.IsProjectInit() {
				err := app.InitProject("")
				if err != nil {
					return err
				}
			}

			err = app.Reconcile()
			if err != nil {
				return err
			}

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen)
			if err != nil {
				return err
			}

			return nil
		},
	}

	startCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "OLAP database driver")
	startCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "OLAP database DSN")
	startCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for the UI and runtime")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for the runtime's gRPC service")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the runtime")
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Disable opening the browser window")

	return startCmd
}
