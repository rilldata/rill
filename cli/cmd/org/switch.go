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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Switching org...")
			sp.Start()

			_, err := dotrill.GetDefaultOrg()
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultOrg(args[0])
			if err != nil {
				return err
			}

			sp.Stop()
			fmt.Printf("Set default organization to %q", args[0])
			return nil
		},
	}

	return switchCmd
}
