package admin

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// AdminCmd represents the admin command
func AdminCmd(ch *cmdutil.Helper) *cobra.Command {
	internalGroupID := ""
	adminCmd := &cobra.Command{
		Use:     "admin",
		Hidden:  !ch.IsDev(),
		Short:   "Manage an admin server",
		GroupID: internalGroupID,
	}
	adminCmd.AddCommand(PingCmd(ch))
	adminCmd.AddCommand(StartCmd(ch))
	adminCmd.AddCommand(SwitchCmd(ch))
	return adminCmd
}
