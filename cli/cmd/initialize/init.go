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
		Use:       "init [example|list]",
		Short:     "Initializing the example project",
		Long:      `Initializing the example project`,
		ValidArgs: []string{"example", "list"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "example":
				err := example.InitExample(projectName, projectDir)
				if err != nil {
					return fmt.Errorf("Example project not found, Project Name:%s, Error:%v", projectName, err)
				}
				fmt.Printf("Example project '%s' unpacked at path '%s'", projectName, projectDir)
			case "list":
				exampleList, err := example.ListExamples()
				if err != nil {
					return fmt.Errorf("Example projects are not available, Error:%v", err)
				}
				fmt.Printf("Available Example project are: %v", exampleList)
			}
			return nil
		},
	}

	initCmd.Flags().StringVarP(&projectName, "name", "p", "default", "Example project Name")
	initCmd.Flags().StringVarP(&projectDir, "dir", "d", ".", "Example project directory")
	return initCmd
}
