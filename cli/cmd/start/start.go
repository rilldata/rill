package start

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/spf13/cobra"
)

// StartCmd represents the start command
func StartCmd(cfg *config.Config) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var httpPort int
	var grpcPort int
	var verbose bool
	var readonly bool
	var noUI bool
	var noOpen bool
	var strict bool
	var logFormat string
	var variables []string
	var exampleName string

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
				if !cfg.Interactive {
					return fmt.Errorf("required arg <path> missing")
				}

				fmt.Println("No existing project found. Enter name to initialize a new Rill project.")
				questions := []*survey.Question{
					{
						Name: "name",
						Prompt: &survey.Input{
							Message: "Enter project name",
							Default: cmdutil.DefaultProjectName(),
						},
						Validate: func(any interface{}) error {
							name := any.(string)
							if name == "" {
								return fmt.Errorf("empty name")
							}
							return nil
						},
					},
				}

				if !cmd.Flags().Changed("example") {
					if err := survey.Ask(questions, &projectPath); err != nil {
						return err
					}
				} else {
					projectPath = exampleName
				}
			}

			parsedLogFormat, ok := local.ParseLogFormat(logFormat)
			if !ok {
				return fmt.Errorf("invalid log format %q", logFormat)
			}

			app, err := local.NewApp(cmd.Context(), cfg.Version, verbose, olapDriver, olapDSN, projectPath, parsedLogFormat, variables)
			if err != nil {
				return err
			}
			defer app.Close()

			if cmd.Flags().Changed("example") {
				if exampleName != "" {
					fmt.Println("Visit our documentation for more examples: https://docs.rilldata.com.")
					fmt.Println("")
				}

				err = app.InitProject(exampleName)
				if err != nil {
					return fmt.Errorf("init project: %w", err)
				}
			}

			err = app.Reconcile(strict)
			if err != nil {
				return fmt.Errorf("reconcile project: %w", err)
			}

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
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for gRPC")
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&strict, "strict", false, "Exit if project has build errors")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	startCmd.Flags().StringSliceVarP(&variables, "env", "e", []string{}, "Set project variables")
	startCmd.Flags().StringVar(&exampleName, "example", "default", "Name of example project")

	return startCmd
}
