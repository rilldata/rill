package superuser

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func SuperuserCmd(cfg *config.Config) *cobra.Command {
	superuserCmd := &cobra.Command{
		Use:   "superuser",
		Short: "Manage super users",
	}

	superuserCmd.AddCommand(ListCmd(cfg))
	superuserCmd.AddCommand(AddCmd(cfg))
	superuserCmd.AddCommand(RemoveCmd(cfg))

	return superuserCmd
}
