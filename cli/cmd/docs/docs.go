package docs

import (
	"log"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/spf13/cobra"
)

var docsUrl = "https://docs.rilldata.com"

// docsCmd represents the docs command
func DocsCmd() *cobra.Command {
	var docsCmd = &cobra.Command{
		Use:   "docs",
		Short: "Show rill docs",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			err := browser.Open(docsUrl)
			if err != nil {
				log.Fatalf("Couldn't open browser: %v", err)
			}
		},
	}
	return docsCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// docsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// docsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
