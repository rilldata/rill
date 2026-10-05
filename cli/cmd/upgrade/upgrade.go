package upgrade

import (
	"fmt"

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
			if !force {
				upToDate, latest, err := isUpToDate(cmd, ch)
				if err == nil && upToDate {
					fmt.Printf("Rill is already up to date (%s). Use --force to reinstall.\n", latest)
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

// isUpToDate reports whether the running CLI version is greater than or equal to the latest released version.
// It returns false (without error) for development builds, where the current version is unknown.
func isUpToDate(cmd *cobra.Command, ch *cmdutil.Helper) (bool, string, error) {
	if ch.Version.Number == "" {
		return false, "", nil
	}

	latest, err := ch.LatestVersion(cmd.Context())
	if err != nil {
		return false, "", err
	}

	upToDate, err := versionUpToDate(ch.Version.Number, latest)
	if err != nil {
		return false, "", err
	}

	// The cached latest version can be stale. Only trust it when it says an upgrade is needed;
	// otherwise refetch from Github so a new release is not missed.
	if upToDate {
		latest, err = ch.RefreshLatestVersion(cmd.Context())
		if err != nil {
			return false, "", err
		}

		upToDate, err = versionUpToDate(ch.Version.Number, latest)
		if err != nil {
			return false, "", err
		}
	}

	return upToDate, latest, nil
}

// versionUpToDate reports whether current is greater than or equal to latest.
func versionUpToDate(current, latest string) (bool, error) {
	currentV, err := goversion.NewVersion(current)
	if err != nil {
		return false, err
	}

	latestV, err := goversion.NewVersion(latest)
	if err != nil {
		return false, err
	}

	return currentV.GreaterThanOrEqual(latestV), nil
}
