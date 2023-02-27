package org

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func OrgCmd(ver version.Version) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:    "org",
		Hidden: !ver.IsDev(),
		Short:  "Manage organisations",
	}
	orgCmd.AddCommand(CreateCmd(ver))
	orgCmd.AddCommand(EditCmd(ver))
	orgCmd.AddCommand(ShowCmd(ver))
	orgCmd.AddCommand(CloseCmd(ver))
	orgCmd.AddCommand(InviteCmd(ver))
	orgCmd.AddCommand(MembersCmd(ver))
	orgCmd.AddCommand(SwitchCmd(ver))

	return orgCmd
}
