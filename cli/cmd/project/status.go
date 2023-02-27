package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func StatusCmd(ver version.Version) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.ExactArgs(1),
		Short: "Status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}
	return statusCmd
}
