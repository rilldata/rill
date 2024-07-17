package runtime

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func RuntimeCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage a runtime",
	}

	cmd.AddCommand(ManagerTokenCmd(ch))
	cmd.AddCommand(ListInstancesCmd(ch))
	cmd.AddCommand(DeleteInstanceCmd(ch))

	return cmd
}
