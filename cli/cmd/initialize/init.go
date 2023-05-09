package initialize

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/examples"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd(cfg *config.Config) *cobra.Command {
	var projectPath string
	var olapDriver string
	var olapDSN string
	var exampleName string
	var listExamples bool
	var verbose bool
	var variables []string

	initCmd := &cobra.Command{
		Use:   "init [<path>]",
		Short: "Initialize a new project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				projectPath = args[0]
				if strings.HasSuffix(projectPath, ".git") {
					repoName, err := gitutil.CloneRepo(projectPath)
					if err != nil {
						return fmt.Errorf("clone repo error: %w", err)
					}

					projectPath = repoName
				}
			} else {
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

				if !listExamples && !cmd.Flags().Changed("example") {
					if err := survey.Ask(questions, &projectPath); err != nil {
						return err
					}
				} else {
					projectPath = exampleName
				}
			}

			// List examples and exit
			if listExamples {
				fmt.Println("The built-in examples are: ")
				names, err := examples.List()
				if err != nil {
					return err
				}
				for _, name := range names {
					fmt.Printf("- %s\n", name)
				}
				fmt.Println("\nVisit our documentation for more examples: https://docs.rilldata.com")
				return nil
			}

			fmt.Println("This application is alpha and we want to hear from you if you have any questions or ideas to share!")
			fmt.Println("You can reach us in our Rill Discord server at https://bit.ly/3NSMKdT.")
			fmt.Println("")

			app, err := local.NewApp(cmd.Context(), cfg.Version, verbose, olapDriver, olapDSN, projectPath, local.LogFormatConsole, variables)
			if err != nil {
				return err
			}
			defer app.Close()

			if app.IsProjectInit() {
				if projectPath == "." {
					return fmt.Errorf("a Rill project already exists in the current directory")
				}
				return fmt.Errorf("a Rill project already exists in directory %q", projectPath)
			}

			// Only use example=default if --example was explicitly set.
			// Otherwise, default to an empty project.
			if !cmd.Flags().Changed("example") {
				exampleName = ""
			}

			if exampleName != "" {
				fmt.Println("Visit our documentation for more examples: https://docs.rilldata.com.")
				fmt.Println("")
			}

			err = app.InitProject(exampleName)
			if err != nil {
				return fmt.Errorf("init project: %w", err)
			}

			err = app.Reconcile(false)
			if err != nil {
				return fmt.Errorf("reconcile project: %w", err)
			}

			return nil
		},
	}

	initCmd.Flags().SortFlags = false
	initCmd.Flags().BoolVar(&listExamples, "list-examples", false, "List available example projects")
	initCmd.Flags().StringVar(&exampleName, "example", "default", "Name of example project")
	initCmd.Flags().Lookup("example").NoOptDefVal = "default" // Allows "--example" without a specific name
	initCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	initCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	initCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	initCmd.Flags().StringSliceVarP(&variables, "env", "e", []string{}, "Set project variables")

	return initCmd
}
