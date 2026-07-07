package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ExtendTrialCmd(ch *cmdutil.Helper) *cobra.Command {
	var org string
	var days int32
	setCmd := &cobra.Command{
		Use:   "extend-trial",
		Short: "Extend trial for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if org == "" {
				return fmt.Errorf("please set --org")
			}

			if days <= 0 || days > 30 {
				return fmt.Errorf("please set --days between 1 and 30")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.SudoExtendTrial(ctx, &adminv1.SudoExtendTrialRequest{
				Org:  org,
				Days: days,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Extended trial for organization %q till %s\n", org, res.TrialEnd.AsTime().Format("2006-01-02"))

			return nil
		},
	}

	setCmd.Flags().SortFlags = false
	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().Int32Var(&days, "days", 0, "Number of days to extend trial")
	return setCmd
}
