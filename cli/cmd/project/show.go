package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ShowCmd(cfg *config.Config) *cobra.Command {
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
