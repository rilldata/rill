package project

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/cmd/project/triggers"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func TriggerCmd(cfg *config.Config) *cobra.Command {
	triggerCmd := &cobra.Command{
		Use:               "trigger",
		Hidden:            !cfg.IsDev(),
		Short:             "Send trigger to deployment",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrg(cfg)),
	}

	triggerCmd.AddCommand(triggers.ReconcileCmd(cfg))
	triggerCmd.AddCommand(triggers.RefreshCmd(cfg))
	triggerCmd.AddCommand(triggers.ResetCmd(cfg))
	return triggerCmd
}
