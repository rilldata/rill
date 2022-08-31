package importsource

import (
	"fmt"

	"github.com/spf13/cobra"
)

// importSourceCmd represents the importSource command
func ImportSourceCmd() *cobra.Command {
	var importSourceCmd = &cobra.Command{
		Use:   "importSource",
		Short: "A brief description of your rill importSource",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			sourcePath, _ := cmd.Flags().GetString("sourcePath")
			if sourcePath != "" {
				fmt.Printf("importSource called with source file Path %s \n", sourcePath)
			} else {
				fmt.Printf("importSource called with source file Path '.' \n")
			}
		},
		ValidArgs: []string{"--sourcePath", "-p"},
	}

	importSourceCmd.Flags().StringP("sourcePath", "p", ".", "Source file path for Rill")

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
