package initialize

import (
	"fmt"

	example "github.com/rilldata/rill/cli/pkg/examples"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd() *cobra.Command {
	var projectName string
	var projectDir string
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializing the example project",
		Long:  `Initializing the example project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := example.InitExample(projectName, projectDir)
			if err != nil {
				fmt.Println("Example project not found, Project Name:", projectName)
				return err
			}

			fmt.Printf("Example project '%s' unpacked at path '%s'", projectName, projectDir)
			return nil
		},
	}

	initCmd.Flags().StringVarP(&projectName, "example", "p", "rill_example", "Example project Name")
	initCmd.Flags().StringVarP(&projectDir, "dir", "d", ".", "Example project directory")
	return initCmd
}
