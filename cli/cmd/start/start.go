package start

import (
	"fmt"

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
		Short: "Build project and start web app",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := local.NewApp(cmd.Context(), ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			// If not initialized, init repo with an empty project
			if !app.IsProjectInit() {
				err := app.InitProject("")
				if err != nil {
					return fmt.Errorf("init project: %w", err)
				}
			}

			err = app.Reconcile()
			if err != nil {
				return fmt.Errorf("reconcile project: %w", err)
			}

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen)
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser")
	startCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	startCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for gRPC")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return startCmd
}
