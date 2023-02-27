package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func ShowCmd(ver version.Version) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return showCmd
}
