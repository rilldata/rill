package embed

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func EmbedCmd(ch *cmdutil.Helper) *cobra.Command {
	embedCmd := &cobra.Command{
		Use:   "embed",
		Short: "Manage embeds",
	}

	embedCmd.AddCommand(OpenCmd(ch))

	return embedCmd
}
