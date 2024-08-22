package start

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
)

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
	var allowedOrigins []string
	var tlsCertPath string
	var tlsKeyPath string

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
			} else if !rillv1beta.HasRillProject(".") {
				if !ch.Interactive {
					return fmt.Errorf("required arg <path> missing")
				}

				currentDir, err := filepath.Abs("")
				if err != nil {
					return err
				}

				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}

				displayPath := currentDir
				defval := true
				if currentDir == homeDir {
					defval = false
					displayPath = "~/"
				} else if strings.HasPrefix(currentDir, homeDir) {
					displayPath = strings.Replace(currentDir, homeDir, "~", 1)
				}

				msg := fmt.Sprintf("Rill will create project files in %q. Do you want to continue?", displayPath)
				confirm, err := cmdutil.ConfirmPrompt(msg, "", defval)
				if err != nil {
					return err
				}
				if !confirm {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			// Default to the current directory if no path is provided
			if projectPath == "" {
				projectPath = "."
			}

			// Check that projectPath doesn't have an excessive number of files.
			// Note: Relies on ListRecursive enforcing drivers.RepoListLimit.
			if _, err := os.Stat(projectPath); err == nil {
				repo, _, err := cmdutil.RepoForProjectPath(projectPath)
				if err != nil {
					return err
				}
				_, err = repo.ListRecursive(cmd.Context(), "**", false)
				if err != nil {
					if errors.Is(err, drivers.ErrRepoListLimitExceeded) {
						ch.PrintfError("The project directory exceeds the limit of %d files. Please open Rill against a directory with fewer files or set \"ignore_paths\" in rill.yaml.\n", drivers.RepoListLimit)
						return nil
					}
					return fmt.Errorf("failed to list project files: %w", err)
				}
			}

			// Parse log format
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
			if (tlsCertPath != "" && tlsKeyPath == "") || (tlsCertPath == "" && tlsKeyPath != "") {
				return fmt.Errorf("both --tls-cert and --tls-key must be provided")
			} else if tlsCertPath != "" && tlsKeyPath != "" {
				// Check to ensure the paths are valid
				if _, err := os.Stat(tlsCertPath); os.IsNotExist(err) {
					return fmt.Errorf("certificate not found: %s", tlsCertPath)
				}
				if _, err := os.Stat(tlsKeyPath); os.IsNotExist(err) {
					return fmt.Errorf("key not found: %s", tlsKeyPath)
				}
			}

			scheme := "http"
			if tlsCertPath != "" && tlsKeyPath != "" {
				scheme = "https"
			}
			localURL := fmt.Sprintf("%s://localhost:%d", scheme, httpPort)

			allowedOrigins = append(allowedOrigins, localURL)

			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Ch:             ch,
				Verbose:        verbose,
				Debug:          debug,
				Reset:          reset,
				Environment:    environment,
				OlapDriver:     olapDriver,
				OlapDSN:        olapDSN,
				ProjectPath:    projectPath,
				LogFormat:      parsedLogFormat,
				Variables:      varsMap,
				LocalURL:       localURL,
				AllowedOrigins: allowedOrigins,
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID, _ := ch.CurrentUserID(cmd.Context())

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen, readonly, userID, tlsCertPath, tlsKeyPath)
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
	startCmd.Flags().StringVar(&tlsCertPath, "tls-cert", "", "Path to TLS certificate")
	startCmd.Flags().StringVar(&tlsKeyPath, "tls-key", "", "Path to TLS key file")

	// --env was previously used for variables, but is now used to set the environment name. We maintain backwards compatibility by keeping --env as a slice var, and setting any value containing an equals sign as a variable.
	startCmd.Flags().StringSliceVarP(&env, "env", "e", []string{}, `Environment name (default "dev")`)
	startCmd.Flags().StringSliceVarP(&vars, "var", "v", []string{}, "Set project variables")
	startCmd.Flags().StringSliceVarP(&allowedOrigins, "allowed-origins", "", []string{}, "add additional allowed-origins")

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
