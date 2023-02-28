package runtime

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// RuntimeCmd represents the runtime command
func RuntimeCmd(ver version.Version) *cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:    "runtime",
		Hidden: !ver.IsDev(),
		Short:  "Manage stand-alone runtimes",
	}
	runtimeCmd.AddCommand(StartCmd(ver))
	return runtimeCmd
}
