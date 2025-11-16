package token

import (
	"sort"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var refreshTokensOnly bool

	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List personal access tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var refresh *bool
			// check if flag was explicitly set
			if cmd.Flags().Changed("refresh") {
				refresh = &refreshTokensOnly
			}

			res, err := client.ListUserAuthTokens(cmd.Context(), &adminv1.ListUserAuthTokensRequest{
				UserId:    "current",
				PageSize:  pageSize,
				PageToken: pageToken,
				Refresh:   refresh,
			})
			if err != nil {
				return err
			}

			if len(res.Tokens) == 0 {
				if refreshTokensOnly {
					ch.PrintfWarn("No refresh tokens found\n")
				} else {
					ch.PrintfWarn("No tokens found\n")
				}
				return nil
			}

			// If the result set is smaller than the page size, sort by creation time (newest first).
			if pageToken == "" && uint32(len(res.Tokens)) < pageSize {
				sort.Slice(res.Tokens, func(i, j int) bool {
					return res.Tokens[i].CreatedOn.AsTime().After(res.Tokens[j].CreatedOn.AsTime())
				})
			}

			ch.PrintUserTokens(res.Tokens)

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}

	listCmd.Flags().Uint32Var(&pageSize, "page-size", 1000, "Number of tokens to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	listCmd.Flags().BoolVar(&refreshTokensOnly, "refresh", false, "List refresh tokens only")

	return listCmd
}
