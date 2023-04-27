package docs

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/spf13/cobra"
)

func DocsCmd(cfg *config.Config, rootCmd *cobra.Command) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:    "docs",
		Hidden: !cfg.IsDev(),
		Short:  "Manage documentation",
	}
	docsCmd.AddCommand(OpenCmd())
	docsCmd.AddCommand(GenerateCmd(rootCmd))

	return docsCmd
}
