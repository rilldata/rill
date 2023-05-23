package user

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func UserCmd(cfg *config.Config) *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage superusers",
	}

	userCmd.AddCommand(SearchCmd(cfg))

	return userCmd
}
