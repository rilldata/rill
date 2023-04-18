package docs

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func GenerateCmd(rootCmd *cobra.Command) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate CLI documentation",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := "docs/docs/references/cmd/"
			if len(args) > 0 {
				dir = args[0]
			}
			err := doc.GenMarkdownTree(rootCmd, dir)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return docsCmd
}
