package sla

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"

	"github.com/spf13/cobra"
)

func SLACmd(ch *cmdutil.Helper) *cobra.Command {
	slaCmd := &cobra.Command{
		Use:   "sla",
		Short: "Manage SLA for project in an organization",
	}

	slaCmd.AddCommand(GetCmd(ch))
	slaCmd.AddCommand(SetCmd(ch))

	return slaCmd
}
