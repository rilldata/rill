package initialize

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/examples"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd(ver string) *cobra.Command {
	var repoDSN string
	var olapDriver string
	var olapDSN string
	var exampleName string
	var listExamples bool

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Rill project",
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

			app, err := local.NewApp(ver, false, olapDriver, olapDSN, repoDSN)
			if err != nil {
				return err
			}

			if app.IsProjectInit() {
				if repoDSN == "." {
					return fmt.Errorf("a Rill project already exists in the current directory")
				} else {
					return fmt.Errorf("a Rill project already exists in directory '%s'", repoDSN)
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

	initCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "OLAP database driver")
	initCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "OLAP database DSN")
	initCmd.Flags().StringVar(&repoDSN, "dir", ".", "Directory to initialize")
	initCmd.Flags().StringVar(&exampleName, "example", "", "Name of example project")
	initCmd.Flags().BoolVar(&listExamples, "list-examples", false, "List available example projects")

	return initCmd
}
