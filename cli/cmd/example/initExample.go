package example

import (
	"fmt"

	"github.com/spf13/cobra"
)

// InitExampleCmd represents the init-example command
func InitExampleCmd() *cobra.Command {
	var initExampleCmd = &cobra.Command{
		Use:   "init-example",
		Short: "A brief description of rill init-example",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("init-example called")
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
