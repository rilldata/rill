package triggers

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func RefreshCmd(cfg *config.Config) *cobra.Command {
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("refresh called")
			return nil
		},
	}

	return refreshCmd
}
