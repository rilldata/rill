package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func EditCmd(ver version.Version) *cobra.Command {
	var token string

	editCmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.ExactArgs(1),
		Short: "Create",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&token, "display-name", "noname", "Display name")

	return editCmd
}
