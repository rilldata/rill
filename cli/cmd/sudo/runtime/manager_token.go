package runtime

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ManagerTokenCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager-token <host>",
		Args:  cobra.ExactArgs(1),
		Short: "Returns a token with full manager permissions for a runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.SudoIssueRuntimeManagerToken(ctx, &adminv1.SudoIssueRuntimeManagerTokenRequest{
				Host: args[0],
			})
			if err != nil {
				return err
			}

			ch.Printf("%s\n", res.Token)
			return nil
		},
	}

	return cmd
}
