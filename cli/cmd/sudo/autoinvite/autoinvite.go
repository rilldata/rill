package autoinvite

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func AutoinviteCmd(cfg *config.Config) *cobra.Command {
	autoinviteCmd := &cobra.Command{
		Use:   "autoinvite",
		Short: "Auto invite users from a domain",
	}
	autoinviteCmd.AddCommand(AddCmd(cfg))
	autoinviteCmd.AddCommand(RemoveCmd(cfg))

	return autoinviteCmd
}
