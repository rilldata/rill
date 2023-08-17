package project

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all the projects",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				cmdutil.PrintlnWarn("No projects found")
				return nil
			}

			cmdutil.PrintlnSuccess("Projects list")
			cmdutil.TablePrinter(toTable(res.Projects))
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
