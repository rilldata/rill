package user

import (
	"context"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var projectName string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				res, err := client.ListProjectMembers(context.Background(), &adminv1.ListProjectMembersRequest{
					Organization: cfg.Org,
					Project:      projectName,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(res.Members)
				cmdutil.PrintInvites(res.Invites)
				// TODO: user groups
			} else {
				res, err := client.ListOrganizationMembers(context.Background(), &adminv1.ListOrganizationMembersRequest{
					Organization: cfg.Org,
				})
				if err != nil {
					return err
				}

				cmdutil.PrintMembers(res.Members)
				cmdutil.PrintInvites(res.Invites)
				// TODO: user groups
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")

	return listCmd
}
