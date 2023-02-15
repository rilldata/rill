package start

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// StartCmd represents the start command
func StartCmd(ver version.Version) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var projectPath string
	var httpPort int
	var grpcPort int
	var verbose bool
	var readonly bool
	var noUI bool
	var noOpen bool
	var strict bool
	var logFormat string

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Build project and start web app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedLogFormat, ok := local.ParseLogFormat(logFormat)
			if !ok {
				return fmt.Errorf("invalid log format %q", logFormat)
			}

			if len(args) > 0 {
				// Clone the given repository to the given directory/repoName
				url := args[0]
				// This can be added as current dir as well instead of repoName
				repoName := url[strings.LastIndex(url, "/")+1:]
				projectPath = repoName

				_, err := git.PlainClone(repoName, false, &git.CloneOptions{
					URL:      url,
					Progress: os.Stdout,
				})
				if err != nil {
					fmt.Println("git clone error:", err)
					projectPath = url
				}
			}

			app, err := local.NewApp(cmd.Context(), ver, verbose, olapDriver, olapDSN, projectPath, parsedLogFormat)
			if err != nil {
				return err
			}
			defer app.Close()

			// If not initialized, init repo with an empty project
			if !app.IsProjectInit() {
				err = app.InitProject("")
				if err != nil {
					return fmt.Errorf("init project: %w", err)
				}
			}

			err = app.Reconcile(strict)
			if err != nil {
				return fmt.Errorf("reconcile project: %w", err)
			}

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen, readonly)
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
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&strict, "strict", false, "Exit if project has build errors")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")

	return startCmd
}
