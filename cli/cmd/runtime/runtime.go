package runtime

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

// RuntimeCmd represents the runtime command
func RuntimeCmd(cfg *config.Config) *cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:    "runtime",
		Hidden: !cfg.IsDev(),
		Short:  "Manage stand-alone runtimes",
	}
	runtimeCmd.AddCommand(StartCmd(cfg))
	runtimeCmd.AddCommand(PingCmd(cfg))
	return runtimeCmd
}
