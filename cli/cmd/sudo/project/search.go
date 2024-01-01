package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SearchCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var sla bool

	searchCmd := &cobra.Command{
		Use:   "search <pattern>",
		Args:  cobra.ExactArgs(1),
		Short: "Search projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: args[0],
				ProdSla:     sla,
				PageSize:    pageSize,
				PageToken:   pageToken,
			})
			if err != nil {
				return err
			}
			if len(res.Names) == 0 {
				ch.Printer.PrintlnWarn("No projects found")
				return nil
			}

			err = ch.Printer.PrintResource(res.Names)
			if err != nil {
				return err
			}

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().BoolVar(&sla, "sla", false, "Show Projects with SLA enabled")
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}
