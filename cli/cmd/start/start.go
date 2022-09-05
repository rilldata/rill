package start

import (
	"fmt"

	"github.com/spf13/cobra"
)

// StartCmd represents the start command
func StartCmd() *cobra.Command {
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "A brief description of rill start",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("start called")
		},
	}

	return startCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
