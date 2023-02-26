package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func EditCmd(ver version.Version) *cobra.Command {
	var displayName string

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.ExactArgs(1),
		Short: "Edit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")

	return editCmd
}
