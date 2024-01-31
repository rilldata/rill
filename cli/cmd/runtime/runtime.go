package runtime

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// RuntimeCmd represents the runtime command
func RuntimeCmd(ch *cmdutil.Helper) *cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:    "runtime",
		Hidden: !ch.Config.IsDev(),
		Short:  "Manage stand-alone runtimes",
	}
	runtimeCmd.AddCommand(StartCmd(ch))
	runtimeCmd.AddCommand(PingCmd(ch))
	runtimeCmd.AddCommand(InstallDuckDBExtensionsCmd(ch))
	return runtimeCmd
}
