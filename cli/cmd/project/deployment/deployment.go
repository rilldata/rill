package deployment

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func DeploymentsCmd(ch *cmdutil.Helper) *cobra.Command {
	deploymentCmd := &cobra.Command{
		Use:   "deployment",
		Short: "Manage project deployments",
	}

	deploymentCmd.AddCommand(ListCmd(ch))
	deploymentCmd.AddCommand(ShowCmd(ch))
	deploymentCmd.AddCommand(CreateCmd(ch))
	deploymentCmd.AddCommand(DeleteCmd(ch))
	deploymentCmd.AddCommand(StartCmd(ch))
	deploymentCmd.AddCommand(StopCmd(ch))
	return deploymentCmd
}
