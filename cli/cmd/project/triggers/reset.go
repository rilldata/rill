package triggers

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/remote"
	"github.com/spf13/cobra"
)

func ResetCmd(cfg *config.Config) *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("reset called")
			adm, err := remote.NewAdminService()
			if err != nil {
				return err
			}

			defer adm.Close()

			err = adm.TriggerRedeploy(cmd.Context(), cfg.Org, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}

	return resetCmd
}
