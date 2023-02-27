package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func DeleteCmd(ver version.Version) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Args:  cobra.ExactArgs(1),
		Short: "Delete",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	return deleteCmd
}
