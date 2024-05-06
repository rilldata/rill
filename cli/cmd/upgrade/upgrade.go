package upgrade

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/installscript"
	"github.com/spf13/cobra"
)

func UpgradeCmd(ch *cmdutil.Helper) *cobra.Command {
	var version string
	var nightly bool

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Rill to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version != "" {
				return installscript.Install(cmd.Context(), version)
			}
			if nightly {
				return installscript.Install(cmd.Context(), "nightly")
			}
			return installscript.Install(cmd.Context(), "")
		},
	}

	upgradeCmd.Flags().StringVar(&version, "version", "", "Install a specific version of Rill")
	upgradeCmd.Flags().BoolVar(&nightly, "nightly", false, "Install the latest nightly build")

	return upgradeCmd
}
