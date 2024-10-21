package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RepairCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repair",
		Args:  cobra.NoArgs,
		Short: "Init billing for orgs missing billing info and puts them on trial plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.SudoTriggerBillingRepair(cmd.Context(), &adminv1.SudoTriggerBillingRepairRequest{})
			if err != nil {
				return err
			}

			ch.Printf("Triggered billing repair for orgs\n")

			return nil
		},
	}

	return cmd
}
