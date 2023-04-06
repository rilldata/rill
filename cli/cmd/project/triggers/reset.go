package triggers

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ResetCmd(cfg *config.Config) *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.TriggerRedeploy(cmd.Context(), &adminv1.TriggerRedeployRequest{OrganizationName: cfg.Org, Name: args[0]})
			if err != nil {
				return err
			}

			fmt.Println("reset done", res)

			return nil
		},
	}

	return resetCmd
}
