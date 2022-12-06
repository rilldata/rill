package initialize

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/examples"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd(ver string) *cobra.Command {
	var projectPath string
	var olapDriver string
	var olapDSN string
	var exampleName string
	var listExamples bool
	var verbose bool

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// List examples and exit
			if listExamples {
				names, err := examples.List()
				if err != nil {
					return err
				}
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			}

			fmt.Println("This application is extremely alpha and we want to hear from you if you have any questions or ideas to share!")
			fmt.Println("You can reach us in our Rill Discord server at https://bit.ly/3NSMKdT.")
			fmt.Println("")

			app, err := local.NewApp(ver, verbose, olapDriver, olapDSN, projectPath)
			if err != nil {
				return err
			}

			if app.IsProjectInit() {
				if projectPath == "." {
					return fmt.Errorf("a Rill project already exists in the current directory")
				} else {
					return fmt.Errorf("a Rill project already exists in directory '%s'", projectPath)
				}
			}

			err = app.InitProject(exampleName)
			if err != nil {
				return err
			}

			err = app.Reconcile()
			if err != nil {
				return err
			}

			return nil
		},
	}

	initCmd.Flags().SortFlags = false
	initCmd.Flags().BoolVar(&listExamples, "list-examples", false, "List available example projects")
	initCmd.Flags().StringVar(&exampleName, "example", "", "Name of example project (default \"empty\")")
	initCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	initCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	initCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	initCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return initCmd
}
