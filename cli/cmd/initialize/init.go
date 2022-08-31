package initialize

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
func InitCmd() *cobra.Command {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "A brief description of your rill init",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, _ := cmd.Flags().GetString("projectDir")
			fmt.Printf("init called with project dir: %s \n", projectDir)
		},
		ValidArgs: []string{"--projectDir", "-p"},
	}

	initCmd.Flags().String("projectDir", "p", "Project Directory for Rill")

	return initCmd
}
func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
