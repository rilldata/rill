package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func SwitchCmd(ver version.Version) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return switchCmd
}
