package org

import (
	"fmt"

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
			sp := cmdutil.Spinner("Switching org...")
			sp.Start()

			err := dotrill.SetDefaultOrg(cfg.GetAdminToken())
			if err != nil {
				return err
			}

			sp.Stop()
			fmt.Println("Default org is set to ~/.rill.")
			return nil
		},
	}

	return switchCmd
}
