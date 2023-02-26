package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func ShowCmd(ver version.Version) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Args:  cobra.ExactArgs(1),
		Short: "Show",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	return showCmd
}
