package start

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/spf13/cobra"
)

// maxProjectFiles is the maximum number of files that can be in a project directory.
// It corresponds to the file watcher limit in runtime/drivers/file/repo.go.
const maxProjectFiles = 1000

// StartCmd represents the start command
func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var httpPort int
	var grpcPort int
	var verbose bool
	var debug bool
	var readonly bool
	var reset bool
	var noUI bool
	var noOpen bool
	var logFormat string
	var env []string
	var vars []string

	startCmd := &cobra.Command{
		Use:   "start [<path>]",
		Short: "Build project and start web app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config

			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
				if strings.HasSuffix(projectPath, ".git") {
					repoName, err := gitutil.CloneRepo(projectPath)
					if err != nil {
						return fmt.Errorf("clone repo error: %w", err)
					}

					projectPath = repoName
				}
			} else if !rillv1beta.HasRillProject("") {
				if !cfg.Interactive {
					return fmt.Errorf("required arg <path> missing")
				}

				currentDir, err := filepath.Abs("")
				if err != nil {
					return err
				}

				projectPath = currentDir
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}

				displayPath := currentDir
				defval := true
				if strings.HasPrefix(currentDir, homeDir) {
					displayPath = strings.Replace(currentDir, homeDir, "~", 1)
					if currentDir == homeDir {
						defval = false
						displayPath = "~/"
					}
				}

				msg := fmt.Sprintf("Rill will create project files in %q. Do you want to continue?", displayPath)
				confirm := cmdutil.ConfirmPrompt(msg, "", defval)
				if !confirm {
					ch.Printer.PrintlnWarn("Aborted")
					return nil
				}
			}

			// Check that projectPath doesn't have an excessive number of files
			n, err := countFilesInDirectory(projectPath)
			if err != nil {
				return err
			}
			if n > maxProjectFiles {
				ch.Printer.PrintlnError(fmt.Sprintf("The project directory exceeds the limit of %d files (found %d files). Please open Rill against a directory with fewer files.", maxProjectFiles, n))
				return nil
			}

			parsedLogFormat, ok := local.ParseLogFormat(logFormat)
			if !ok {
				return fmt.Errorf("invalid log format %q", logFormat)
			}

			// Backwards compatibility for --env (see comment on the flag definition for details)
			environment := "dev"
			for _, v := range env {
				if strings.Contains(v, "=") {
					vars = append(vars, v)
				} else {
					environment = v
				}
			}

			client := activity.NewNoopClient()

			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Version:     cfg.Version,
				Verbose:     verbose,
				Debug:       debug,
				Reset:       reset,
				Environment: environment,
				OlapDriver:  olapDriver,
				OlapDSN:     olapDSN,
				ProjectPath: projectPath,
				LogFormat:   parsedLogFormat,
				Variables:   vars,
				Activity:    client,
				AdminURL:    cfg.AdminURL,
				AdminToken:  cfg.AdminToken(),
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID := ""
			if cfg.IsAuthenticated() {
				userID, _ = cmdutil.FetchUserID(context.Background(), cfg)
			}

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen, readonly, userID)
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser")
	startCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	startCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 49009, "Port for gRPC (internal)")
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	startCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")

	// --env was previously used for variables, but is now used to set the environment name. We maintain backwards compatibility by keeping --env as a slice var, and setting any value containing an equals sign as a variable.
	startCmd.Flags().StringSliceVarP(&env, "env", "e", []string{}, `Environment name (default "dev")`)
	startCmd.Flags().StringSliceVarP(&vars, "var", "v", []string{}, "Set project variables")

	return startCmd
}

func countFilesInDirectory(path string) (int, error) {
	var fileCount int

	if path == "" {
		path = "."
	}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	return fileCount, nil
}
