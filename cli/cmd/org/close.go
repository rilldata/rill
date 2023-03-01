package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func CloseCmd(cfg *config.Config) *cobra.Command {
	closeCmd := &cobra.Command{
		Use:   "close",
		Short: "Close",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return closeCmd
}
