package plan

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List plans",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.ListPublicBillingPlans(ctx, &adminv1.ListPublicBillingPlansRequest{})
			if err != nil {
				return err
			}

			if len(resp.Plans) == 0 {
				ch.PrintfWarn("No plans found\n")
				return nil
			}

			ch.PrintPlans(resp.Plans)
			return nil
		},
	}

	return listCmd
}
