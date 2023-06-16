package quota

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func QuotaCmd(cfg *config.Config) *cobra.Command {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage quota for user and org",
	}

	quotaCmd.AddCommand(GetCmd(cfg))
	quotaCmd.AddCommand(SetCmd(cfg))

	return quotaCmd
}
