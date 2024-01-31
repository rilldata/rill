package quota

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func QuotaCmd(ch *cmdutil.Helper) *cobra.Command {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage quota for user and org",
	}

	quotaCmd.AddCommand(GetCmd(ch))
	quotaCmd.AddCommand(SetCmd(ch))

	return quotaCmd
}
