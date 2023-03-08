package admin

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

// AdminCmd represents the admin command
func AdminCmd(cfg *config.Config) *cobra.Command {
	adminCmd := &cobra.Command{
		Use:    "admin",
		Hidden: !cfg.IsDev(),
		Short:  "Manage an admin server",
	}
	adminCmd.AddCommand(StartCmd(cfg))
	adminCmd.AddCommand(PingCmd(cfg))
	return adminCmd
}
