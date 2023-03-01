package org

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func OrgCmd(cfg *config.Config) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:    "org",
		Hidden: !cfg.IsDev(),
		Short:  "Manage organisations",
	}
	orgCmd.AddCommand(CreateCmd(cfg))
	orgCmd.AddCommand(EditCmd(cfg))
	orgCmd.AddCommand(ShowCmd(cfg))
	orgCmd.AddCommand(CloseCmd(cfg))
	orgCmd.AddCommand(InviteCmd(cfg))
	orgCmd.AddCommand(MembersCmd(cfg))
	orgCmd.AddCommand(SwitchCmd(cfg))

	return orgCmd
}
