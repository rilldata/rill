package user

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func UserCmd(cfg *config.Config) *cobra.Command {
	userCmd := &cobra.Command{
		Use:    "user",
		Hidden: !cfg.IsDev(),
		Short:  "Manage users",
	}

	userCmd.AddCommand(ListCmd(cfg))
	userCmd.AddCommand(AddCmd(cfg))
	userCmd.AddCommand(RemoveCmd(cfg))
	userCmd.AddCommand(SetRoleCmd(cfg))

	return userCmd
}
