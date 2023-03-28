package triggers

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ResetCmd(cfg *config.Config) *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("reset called")
			return nil
		},
	}

	return resetCmd
}
