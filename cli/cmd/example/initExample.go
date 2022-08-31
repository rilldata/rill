package example

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initExampleCmd represents the initExample command
func InitExampleCmd() *cobra.Command {
	var initExampleCmd = &cobra.Command{
		Use:   "initExample",
		Short: "A brief description of your rill initExample",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("initExample called")
		},
	}

	return initExampleCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initExampleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initExampleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
