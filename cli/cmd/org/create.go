package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func CreateCmd(ver version.Version) *cobra.Command {
	var displayName string

	createCmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.ExactArgs(1),
		Short: "Create",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")

	return createCmd
}
