package token

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func IssueCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string
	issueCmd := &cobra.Command{
		Use:   "issue [<service>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Issue service token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("service name is required. Use --service flag or provide as an argument")
			}

			res, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				Org:         ch.Org,
				ServiceName: name,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Issued token: %v\n", res.Token)

			return nil
		},
	}

	issueCmd.Flags().SortFlags = false
	issueCmd.Flags().StringVar(&name, "service", "", "Service Name")

	return issueCmd
}
