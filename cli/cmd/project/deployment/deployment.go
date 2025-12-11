package deployment

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func DeploymentsCmd(ch *cmdutil.Helper) *cobra.Command {
	deploymentCmd := &cobra.Command{
		Use:   "deployments",
		Short: "Manage project deployments",
	}

	deploymentCmd.AddCommand(DeploymentsListCmd(ch))
	deploymentCmd.AddCommand(DeploymentShowCmd(ch))
	deploymentCmd.AddCommand(DeploymentCreateCmd(ch))
	deploymentCmd.AddCommand(DeploymentDeleteCmd(ch))
	deploymentCmd.AddCommand(DeploymentStartCmd(ch))
	deploymentCmd.AddCommand(DeploymentStopCmd(ch))

	return deploymentCmd
}
