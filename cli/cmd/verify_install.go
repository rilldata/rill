package cmd

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/spf13/cobra"
)

func verifyInstallCmd(ch *cmdutil.Helper) *cobra.Command {
	internalGroupID := ""
	cmd := &cobra.Command{
		Use:    "verify-install",
		Short:  "Verify installation (called by install script)",
		Hidden: !ch.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ch.Telemetry(cmd.Context()).RecordBehavioralLegacy(activity.BehavioralEventInstallSuccess)
			return nil
		},
		GroupID: internalGroupID,
	}

	return cmd
}
