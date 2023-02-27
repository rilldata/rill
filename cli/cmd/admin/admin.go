package admin

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// AdminCmd represents the admin command
func AdminCmd(ver version.Version) *cobra.Command {
	adminCmd := &cobra.Command{
		Use:    "admin",
		Hidden: !ver.IsDev(),
		Short:  "Manage an admin server",
	}
	adminCmd.AddCommand(StartCmd(ver))
	return adminCmd
}
