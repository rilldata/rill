package info

import (
	"fmt"

	"github.com/spf13/cobra"
)

// InfoCmd represents the info command
func InfoCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "A brief description of rill info",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("info called")
		},
	}

	return cmd
}
func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
