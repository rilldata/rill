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
	userCmd.AddCommand(ListCmd(ch))
	userCmd.AddCommand(ShowCmd(ch))
	userCmd.AddCommand(CreateCmd(ch))
	userCmd.AddCommand(RenameCmd(ch))
	userCmd.AddCommand(EditCmd(ch))
	userCmd.AddCommand(DeleteCmd(ch))

	// Manage user group roles
	userCmd.AddCommand(AddCmd(ch))
	userCmd.AddCommand(SetRoleCmd(ch))
	userCmd.AddCommand(RemoveCmd(ch))

	return userCmd
}

var usergroupRoles = []string{"admin", "viewer"}
