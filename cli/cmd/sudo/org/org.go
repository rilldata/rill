package org

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func OrgCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "org",
		Short:             "Org management for support users",
		PersistentPreRunE: cmdutil.CheckAuth(ch),
	}

	cmd.AddCommand(ShowCmd(ch))
	cmd.AddCommand(JoinCmd(ch))
	cmd.AddCommand(ListAdminsCmd(ch))
	cmd.AddCommand(SetCustomDomainCmd(ch))
	cmd.AddCommand(SetInternalPlanCmd(ch))

	return cmd
}
