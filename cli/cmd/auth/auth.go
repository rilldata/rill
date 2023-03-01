package auth

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func AuthCmd(cfg *config.Config) *cobra.Command {
	authCmd := &cobra.Command{
		Use:    "auth",
		Hidden: !cfg.IsDev(),
		Short:  "Manage authentication",
	}
	authCmd.AddCommand(LoginCmd(cfg))
	authCmd.AddCommand(LogoutCmd(cfg))

	return authCmd
}
