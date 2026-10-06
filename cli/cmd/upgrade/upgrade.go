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

				return installscript.Install(cmd.Context(), version)
			}

			if nightly {
				return installscript.Install(cmd.Context(), "nightly")
			}

			// Skip the download if the current version is already the latest release.
			if !force && !ch.IsDev() {
				latest, err := ch.RefreshLatestVersion(cmd.Context())
				if err != nil {
					ch.PrintfWarn("Could not check latest version: %v\n", err)
				} else if upToDate, err := versionUpToDate(ch.Version.Number, latest); err == nil && upToDate {
					ch.Printf("Rill is already up to date (%s). Use --force to reinstall.\n", latest)
					return nil
				}
			}

			return installscript.Install(cmd.Context(), "")
		},
	}

	upgradeCmd.Flags().StringVar(&version, "version", "", "Install a specific version of Rill")
	upgradeCmd.Flags().BoolVar(&nightly, "nightly", false, "Install the latest nightly build")
	upgradeCmd.Flags().BoolVar(&force, "force", false, "Reinstall even if the current version is already the latest")

	return upgradeCmd
}

// versionUpToDate reports whether current is greater than or equal to latest.
// Prerelease builds (e.g. nightly) are never up to date, so upgrading moves them to the latest stable release.
func versionUpToDate(current, latest string) (bool, error) {
	currentV, err := goversion.NewVersion(current)
	if err != nil {
		return false, err
	}
	if currentV.Prerelease() != "" {
		return false, nil
	}

	latestV, err := goversion.NewVersion(latest)
	if err != nil {
		return false, err
	}

	return currentV.GreaterThanOrEqual(latestV), nil
}
