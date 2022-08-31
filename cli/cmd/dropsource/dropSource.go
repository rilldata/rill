package dropsource

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dropSourceCmd represents the dropSource command
func DropSourceCmd() *cobra.Command {
	var dropSourceCmd = &cobra.Command{
		Use:   "dropSource",
		Short: "A brief description of your rill dropSource",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			sourceName, _ := cmd.Flags().GetString("sourceName")
			if sourceName != "" {
				fmt.Printf("dropSource called with source name: %s \n", sourceName)
			} else {
				fmt.Printf("dropSource called with source name: public \n")
			}
		},
		ValidArgs: []string{"--sourceName", "-n"},
	}

	dropSourceCmd.Flags().StringP("sourceName", "n", "public", "Source Name to be dropped")

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
