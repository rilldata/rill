package user

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				members, err := client.ListProjectMembers(ctx, &adminv1.ListProjectMembersRequest{
					Organization: cfg.Org,
					Project:      projectName,
					PageSize:     pageSize,
					PageToken:    pageToken,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(members.Members)
				if members.NextPageToken != "" {
					cmd.Println()
					cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
				}
			} else {
				members, err := client.ListOrganizationMembers(context.Background(), &adminv1.ListOrganizationMembersRequest{
					Organization: cfg.Org,
					PageSize:     pageSize,
					PageToken:    pageToken,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(members.Members)
				if members.NextPageToken != "" {
					cmd.Println()
					cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
				}
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")
	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of users to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}
