package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := dotrill.SetDefaultOrg(cfg.AdminToken)
			if err != nil {
				return err
			}

			fmt.Println("Default org is set to ~/.rill.")
			return nil
		},
	}

	return switchCmd
}
