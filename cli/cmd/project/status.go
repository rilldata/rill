package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func StatusCmd(cfg *config.Config) *cobra.Command {
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
