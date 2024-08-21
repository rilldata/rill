package org

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func OrgCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Org management for support users",
	}

	cmd.AddCommand(SetCustomDomainCmd(ch))

	return cmd
}
