package importsource

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ImportSourceCmd represents the import-source command
func ImportSourceCmd() *cobra.Command {
	var importSourceCmd = &cobra.Command{
		Use:   "import-source",
		Short: "A brief description of rill import-source",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			sourcePath, _ := cmd.Flags().GetString("source-path")
			if sourcePath != "" {
				fmt.Printf("import-source called with source path %s\n", sourcePath)
			} else {
				fmt.Printf("import-source called with source path '.'\n")
			}
		},
		ValidArgs: []string{"--source-path", "-p"},
	}

	importSourceCmd.Flags().StringP("source-path", "p", ".", "Source path for Rill")

	return importSourceCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importSourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importSourceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
