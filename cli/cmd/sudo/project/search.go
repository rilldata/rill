package project

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
		Use:   "search",
		Args:  cobra.ExactArgs(1),
		Short: "Search projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: args[0],
				PageSize:    pageSize,
				PageToken:   pageToken,
			})
			if err != nil {
				return err
			}
			if len(res.Names) == 0 {
				cmdutil.PrintlnWarn("No projects found")
				return nil
			}

			cmdutil.TablePrinter(res.Names)
			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}
