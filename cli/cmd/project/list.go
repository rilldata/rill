package project

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all the projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListProjectsForOrganization(context.Background(), &adminv1.ListProjectsForOrganizationRequest{
				OrganizationName: cfg.Org,
				PageSize:         pageSize,
				PageToken:        pageToken,
			})
			if err != nil {
				return err
			}

			if len(res.Projects) == 0 {
				ch.Printer.PrintlnWarn("No projects found")
				return nil
			}

			ch.Printer.PrintlnSuccess("Projects list")
			err = ch.Printer.PrintResource(toTable(res.Projects))
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

	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}
