package docs

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func DocsCmd(cfg *config.Config, rootCmd *cobra.Command) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:    "docs",
		Hidden: !cfg.IsDev(),
		Short:  "Manage documentation",
	}
	orgCmd.AddCommand(OpenCmd())
	orgCmd.AddCommand(GenerateCmd(rootCmd))

	return orgCmd
}
