package org

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.GetSpinner(4, "Switching org...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)

			err := dotrill.SetDefaultOrg(cfg.GetAdminToken())
			if err != nil {
				return err
			}

			fmt.Println("Default org is set to ~/.rill.")
			sp.Stop()
			return nil
		},
	}

	return switchCmd
}
