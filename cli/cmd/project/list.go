package project

import (
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
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.ListProjectsForOrganization(cmd.Context(), &adminv1.ListProjectsForOrganizationRequest{
				Org:       ch.Org,
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			if len(res.Projects) == 0 {
				ch.PrintfWarn("No projects found\n")
				return nil
			}

			ch.PrintfSuccess("Projects list\n")
			ch.PrintProjects(res.Projects)

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
