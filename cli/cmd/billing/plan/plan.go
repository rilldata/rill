package plan

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func PlanCmd(ch *cmdutil.Helper) *cobra.Command {
	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Get billing plans",
	}

	planCmd.AddCommand(ListCmd(ch))
	return planCmd
}
