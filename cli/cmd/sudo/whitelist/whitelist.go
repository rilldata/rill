package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func WhitelistCmd(ch *cmdutil.Helper) *cobra.Command {
	whitelistCmd := &cobra.Command{
		Use:   "whitelist",
		Short: "Whitelist users from an email domain",
	}
	whitelistCmd.AddCommand(AddCmd(ch))
	whitelistCmd.AddCommand(RemoveCmd(ch))

	return whitelistCmd
}
