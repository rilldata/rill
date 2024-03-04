package start

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
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
	var certPath string
	var keyPath string

	startCmd := &cobra.Command{
		Use:   "start [<path>]",
		Short: "Build project and start web app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				if !ch.Interactive {
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
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			// Check that projectPath doesn't have an excessive number of files
			n, err := countFilesInDirectory(projectPath)
			if err != nil {
				return err
			}
			if n > maxProjectFiles {
				ch.PrintfError("The project directory exceeds the limit of %d files (found %d files). Please open Rill against a directory with fewer files.\n", maxProjectFiles, n)
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

			// Parser variables from "a=b" format to map
			varsMap, err := parseVariables(vars)
			if err != nil {
				return err
			}

			// If keypath or certpath provided, but not the other, display error
			// If keypath and certpath provided, check if the file exists
			if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
				return fmt.Errorf("both --cert and --key must be provided")
			} else if certPath != "" && keyPath != "" {
				if _, err := os.Stat(certPath); os.IsNotExist(err) {
					return fmt.Errorf("certificate not found: %s", certPath)
				}
				if _, err := os.Stat(keyPath); os.IsNotExist(err) {
					return fmt.Errorf("key not found: %s", keyPath)
				}
			}

			client := activity.NewNoopClient()

			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Version:     ch.Version,
				Verbose:     verbose,
				Debug:       debug,
				Reset:       reset,
				Environment: environment,
				OlapDriver:  olapDriver,
				OlapDSN:     olapDSN,
				ProjectPath: projectPath,
				LogFormat:   parsedLogFormat,
				Variables:   varsMap,
				Activity:    client,
				AdminURL:    ch.AdminURL,
				AdminToken:  ch.AdminToken(),
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID := ""
			if ch.IsAuthenticated() {
				user, _ := ch.CurrentUser(cmd.Context())
				if user != nil {
					userID = user.Id
				}
			}

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen, readonly, userID, certPath, keyPath)
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 49009, "Port for gRPC (internal)")
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	startCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	startCmd.Flags().StringVar(&certPath, "cert", "", "Path to TLS certificate")
	startCmd.Flags().StringVar(&keyPath, "key", "", "Path to TLS key")

	// --env was previously used for variables, but is now used to set the environment name. We maintain backwards compatibility by keeping --env as a slice var, and setting any value containing an equals sign as a variable.
	startCmd.Flags().StringSliceVarP(&env, "env", "e", []string{}, `Environment name (default "dev")`)
	startCmd.Flags().StringSliceVarP(&vars, "var", "v", []string{}, "Set project variables")

	// We have deprecated the ability configure the OLAP database via the CLI. This should now be done via rill.yaml.
	// Keeping these for backwards compatibility for a while.
	startCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	startCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	if err := startCmd.Flags().MarkHidden("db"); err != nil {
		panic(err)
	}
	if err := startCmd.Flags().MarkHidden("db-driver"); err != nil {
		panic(err)
	}

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

func parseVariables(vals []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, v := range vals {
		v, err := godotenv.Unmarshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse variable %q: %w", v, err)
		}
		for k, v := range v {
			res[k] = v
		}
	}
	return res, nil
}
