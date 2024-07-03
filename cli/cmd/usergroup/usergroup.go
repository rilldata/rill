package usergroup

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func UsergroupCmd(ch *cmdutil.Helper) *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "usergroup",
		Short:             "Manage user groups",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	// Manage user groups
	userCmd.AddCommand(CreateCmd(ch))
	userCmd.AddCommand(ShowCmd(ch))
	userCmd.AddCommand(ListCmd(ch))
	userCmd.AddCommand(RemoveCmd(ch))
	// Manage user group roles
	userCmd.AddCommand(SetRoleCmd(ch))
	userCmd.AddCommand(RevokeRoleCmd(ch))
	// Manage user group members
	userCmd.AddCommand(AddUserCmd(ch))
	userCmd.AddCommand(ListUserCmd(ch))
	userCmd.AddCommand(RemoveUserCmd(ch))

	return userCmd
}

var usergroupRoles = []string{"admin", "viewer"}

