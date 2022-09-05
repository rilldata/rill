package dropsource

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DropSourceCmd represents the drop-source command
func DropSourceCmd() *cobra.Command {
	var dropSourceCmd = &cobra.Command{
		Use:   "drop-source",
		Short: "A brief description of rill drop-source",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			sourceName, _ := cmd.Flags().GetString("source-name")
			if sourceName != "" {
				fmt.Printf("drop-source called with source name: %s \n", sourceName)
			} else {
				fmt.Printf("drop-source called with source name: public \n")
			}
		},
		ValidArgs: []string{"--source-name", "-n"},
	}

	dropSourceCmd.Flags().StringP("source-name", "n", "public", "Source Name to be dropped")

	return dropSourceCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dropSourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dropSourceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
