package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func WhitelistCmd(cfg *config.Config) *cobra.Command {
	whitelistCmd := &cobra.Command{
		Use:   "whitelist",
		Short: "Whitelist users from an email domain",
	}
	whitelistCmd.AddCommand(AddCmd(cfg))
	whitelistCmd.AddCommand(RemoveCmd(cfg))

	return whitelistCmd
}
