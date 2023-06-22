package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SearchCmd(cfg *config.Config) *cobra.Command {
	var pageSize uint32
	var pageToken string

	searchCmd := &cobra.Command{
		Use:   "search <email-pattern>",
		Args:  cobra.ExactArgs(1),
		Short: "Search users by email pattern (use % as wildcard)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.SearchUsers(ctx, &adminv1.SearchUsersRequest{
				EmailPattern: args[0],
				PageSize:     pageSize,
				PageToken:    pageToken,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintUsers(res.Users)

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}

	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of users to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}
