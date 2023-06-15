package user

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func UserCmd(cfg *config.Config) *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	userCmd.AddCommand(SearchCmd(cfg))
	userCmd.AddCommand(AssumeCmd(cfg))
	userCmd.AddCommand(UnassumeCmd(cfg))
	userCmd.AddCommand(OpenCmd(cfg))

	return userCmd
}
