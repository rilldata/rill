package user

import (
	"github.com/rilldata/rill/cli/cmd/user/whitelist"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func UserCmd(ch *cmdutil.Helper) *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "user",
		Short:             "Manage users",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	userCmd.AddCommand(ListCmd(ch))
	userCmd.AddCommand(ShowCmd(ch))
	userCmd.AddCommand(AddCmd(ch))
	userCmd.AddCommand(RemoveCmd(ch))
	userCmd.AddCommand(SetRoleCmd(ch))
	userCmd.AddCommand(SetAttributesCmd(ch))
	userCmd.AddCommand(whitelist.WhitelistCmd(ch))

	return userCmd
}

var orgRoles = []string{"admin", "editor", "viewer", "guest"}

var projectRoles = []string{"admin", "editor", "viewer"}
