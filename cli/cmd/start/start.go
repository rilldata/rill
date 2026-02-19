package start

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/envdetect"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// StartCmd represents the start command
func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	var httpPort int
	var grpcPort int
	var verbose bool
	var debug bool
	var readonly bool
	var reset bool
	var pullEnv bool
	var noUI bool
	var noOpen bool
	var previewLocked bool
	var logFormat string
	var envVars, envVarsOld []string
	var environment string
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
				var err error
				projectPath, err = ResolveProjectPath(args[0])
				if err != nil {
					return err
				}
			} else if !cmdutil.HasRillProject(".") {
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

				if currentDir == homeDir {
					confirm, err := cmdutil.ConfirmPrompt(
						"You are trying to start Rill in your home directory, which is not recommended. Are you sure you want to continue?",
						"", false,
					)
					if err != nil {
						return err
					}
					if !confirm {
						ch.PrintfWarn("Aborted\n")
						return nil
					}
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

			// Check for WSL Windows partition usage (based on the project path)
			if envdetect.IsWSLWindowsPartition(projectPath) {
				ch.PrintfWarn("%s\n", envdetect.GetWSLWarningMessage())
				confirm, err := cmdutil.ConfirmPrompt(
					"Do you want to continue anyway?",
					"", false, // Default to "No"
				)
				if err != nil {
					return err
				}
				if !confirm {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			// Parser variables from "a=b" format to map
			envVars = append(envVars, envVarsOld...)
			envVarsMap, err := ParseVariables(envVars)
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

			ch.Interactive = false // Disable interactive mode for the app server
			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Ch:             ch,
				Verbose:        verbose,
				Debug:          debug,
				Reset:          reset,
				PullEnv:        pullEnv,
				Environment:    environment,
				ProjectPath:    projectPath,
				LogFormat:      logFormat,
				Variables:      envVarsMap,
				LocalURL:       localURL,
				AllowedOrigins: allowedOrigins,
				ServeUI:        !noUI,
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID, _ := ch.CurrentUserID(cmd.Context())

			err = app.Serve(local.ServeOptions{
				HTTPPort:      httpPort,
				GRPCPort:      grpcPort,
				EnableUI:      !noUI,
				OpenBrowser:   !noOpen,
				Readonly:      readonly,
				PreviewLocked: previewLocked,
				UserID:        userID,
				TLSCertPath:   tlsCertPath,
				TLSKeyPath:    tlsKeyPath,
			})
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Set environment variables")
	startCmd.Flags().StringVar(&environment, "environment", "dev", `Environment name`)
	startCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	startCmd.Flags().BoolVar(&pullEnv, "pull-env", true, "Pull environment variables from Rill Cloud before starting the project")
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&previewLocked, "preview-locked", false, "Run in locked preview mode (no mode toggle)")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 49009, "Port for gRPC (internal)")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	startCmd.Flags().StringVar(&tlsCertPath, "tls-cert", "", "Path to TLS certificate")
	startCmd.Flags().StringVar(&tlsKeyPath, "tls-key", "", "Path to TLS key file")
	startCmd.Flags().StringSliceVarP(&allowedOrigins, "allowed-origins", "", []string{}, "Override allowed origins for CORS")

	// Deprecated support for "--var": replaced by "--env".
	startCmd.Flags().StringSliceVarP(&envVarsOld, "var", "v", []string{}, "Set environment variables")
	if err := startCmd.Flags().MarkHidden("var"); err != nil {
		panic(err)
	}

	// Deprecated support for "--readonly". Projects should be shared via Rill Cloud.
	if err := startCmd.Flags().MarkHidden("readonly"); err != nil {
		panic(err)
	}

	return startCmd
}

func ResolveProjectPath(path string) (string, error) {
	if strings.HasSuffix(path, ".git") {
		repoName, err := gitutil.CloneRepo(path)
		if err != nil {
			return "", fmt.Errorf("clone repo error: %w", err)
		}
		path = repoName
	}

	return path, nil
}

func ParseVariables(vals []string) (map[string]string, error) {
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
