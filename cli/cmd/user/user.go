package user

import (
	"github.com/rilldata/rill/cli/cmd/user/whitelist"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func UserCmd(ch *cmdutil.Helper) *cobra.Command {
	cfg := ch.Config
	userCmd := &cobra.Command{
		Use:               "user",
		Short:             "Manage users",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	userCmd.AddCommand(ListCmd(ch))
	userCmd.AddCommand(AddCmd(ch))
	userCmd.AddCommand(RemoveCmd(ch))
	userCmd.AddCommand(SetRoleCmd(ch))
	userCmd.AddCommand(whitelist.WhitelistCmd(ch))

	return userCmd
}

var userRoles = []string{"admin", "viewer"}
