package runtime

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// RuntimeCmd represents the runtime command
func RuntimeCmd(ch *cmdutil.Helper) *cobra.Command {
	internalGroupID := ""
	runtimeCmd := &cobra.Command{
		Use:     "runtime",
		Hidden:  !ch.IsDev(),
		Short:   "Manage stand-alone runtimes",
		GroupID: internalGroupID,
	}
	runtimeCmd.AddCommand(StartCmd(ch))
	runtimeCmd.AddCommand(PingCmd(ch))
	runtimeCmd.AddCommand(InstallDuckDBExtensionsCmd(ch))
	return runtimeCmd
}
