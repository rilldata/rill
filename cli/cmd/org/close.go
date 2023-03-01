package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func CloseCmd(ver version.Version) *cobra.Command {
	closeCmd := &cobra.Command{
		Use:   "close",
		Short: "Close",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return closeCmd
}
