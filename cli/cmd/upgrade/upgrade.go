package upgrade

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/installscript"
	"github.com/spf13/cobra"
)

func UpgradeCmd(ch *cmdutil.Helper) *cobra.Command {
	var version string
	var nightly bool
	var force bool

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Rill to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version != "" {
				// Parse the version into the canonical form
				v, err := goversion.NewVersion(version)
				if err != nil {
					return err
				}
				version = "v" + v.String()

				return installscript.Install(cmd.Context(), version, force)
			}

			if nightly {
				return installscript.Install(cmd.Context(), "nightly", force)
			}

			return installscript.Install(cmd.Context(), "", force)
		},
	}

	upgradeCmd.Flags().StringVar(&version, "version", "", "Install a specific version of Rill")
	upgradeCmd.Flags().BoolVar(&nightly, "nightly", false, "Install the latest nightly build")
	upgradeCmd.Flags().BoolVar(&force, "force", false, "Reinstall even if the latest version is already installed")

	return upgradeCmd
}
