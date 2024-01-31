package upgrade

import (
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const installScriptURL = "https://cdn.rilldata.com/install.sh"

func UpgradeCmd(ch *cmdutil.Helper) *cobra.Command {
	var nightly bool

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Rill to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Download the install script to a temporary file
			f, err := os.CreateTemp("", "install*.sh")
			if err != nil {
				return err
			}
			defer os.Remove(f.Name())

			resp, err := http.Get(installScriptURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			_, err = io.Copy(f, resp.Body)
			if err != nil {
				return err
			}
			f.Close()

			// Run the install script with bash
			args := []string{f.Name()}
			if nightly {
				args = append(args, "--nightly")
			}
			cmd := exec.Command("/bin/bash", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}

	upgradeCmd.Flags().BoolVar(&nightly, "nightly", false, "Install the latest nightly build")

	return upgradeCmd
}
