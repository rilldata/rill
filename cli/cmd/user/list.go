package user

import (
	"context"
	"strings"

	adminclient "github.com/rilldata/rill/admin/client"
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
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName != "" {
				if strings.HasPrefix(pageToken, "usr") {
					err = listProjectMembers(cmd, client, cfg.Org, projectName, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, "inv") {
					err = listProjectInvites(cmd, client, cfg.Org, projectName, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				} else {
					err = listProjectMembers(cmd, client, cfg.Org, projectName, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}

					err = listProjectInvites(cmd, client, cfg.Org, projectName, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				}

				// TODO: user groups
			} else {
				if strings.HasPrefix(pageToken, "usr") {
					err = listOrgMembers(cmd, client, cfg.Org, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, "inv") {
					err = listOrgInvites(cmd, client, cfg.Org, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				} else {
					err = listOrgMembers(cmd, client, cfg.Org, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}

					err = listOrgInvites(cmd, client, cfg.Org, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				}

				// TODO: user groups
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

func listProjectMembers(cmd *cobra.Command, client *adminclient.Client, org, project, pageToken string, pageSize uint32) error {
	members, err := client.ListProjectMembers(context.Background(), &adminv1.ListProjectMembersRequest{
		Organization: org,
		Project:      project,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintMembers(members.Members)
	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token for members: usr%s\n", members.NextPageToken)
	}

	return nil
}

func listProjectInvites(cmd *cobra.Command, client *adminclient.Client, org, project, pageToken string, pageSize uint32) error {
	invites, err := client.ListProjectInvites(context.Background(), &adminv1.ListProjectInvitesRequest{
		Organization: org,
		Project:      project,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintInvites(invites.Invites)
	if invites.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token for invites: inv%s\n", invites.NextPageToken)
	}

	return nil
}

func listOrgMembers(cmd *cobra.Command, client *adminclient.Client, org, pageToken string, pageSize uint32) error {
	members, err := client.ListOrganizationMembers(context.Background(), &adminv1.ListOrganizationMembersRequest{
		Organization: org,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintMembers(members.Members)
	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token for members: usr%s\n", members.NextPageToken)
	}
	return nil
}

func listOrgInvites(cmd *cobra.Command, client *adminclient.Client, org, pageToken string, pageSize uint32) error {
	invites, err := client.ListOrganizationInvites(context.Background(), &adminv1.ListOrganizationInvitesRequest{
		Organization: org,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintInvites(invites.Invites)
	if invites.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token for invites: inv%s\n", invites.NextPageToken)
	}

	return nil
}
