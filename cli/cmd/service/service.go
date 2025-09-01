package service

import (
	"github.com/rilldata/rill/cli/cmd/service/token"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func ServiceCmd(ch *cmdutil.Helper) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:               "service",
		Short:             "Manage service accounts",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	serviceCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	serviceCmd.AddCommand(ListCmd(ch))
	serviceCmd.AddCommand(CreateCmd(ch))
	serviceCmd.AddCommand(ShowCmd(ch))
	serviceCmd.AddCommand(EditCmd(ch))
	serviceCmd.AddCommand(SetRoleCmd(ch))
	serviceCmd.AddCommand(RemoveCmd(ch))
	serviceCmd.AddCommand(DeleteCmd(ch))
	serviceCmd.AddCommand(token.TokenCmd(ch))

	return serviceCmd
}

var orgRoles = []string{"admin", "editor", "viewer", "guest"}

var projectRoles = []string{"admin", "editor", "viewer"}
