package user

import (
	"context"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectName != "" {
				if strings.HasPrefix(pageToken, "usr") {
					err := listProjectMembers(cmd, ch, ch.Org, projectName, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, "inv") {
					err := listProjectInvites(cmd, ch, ch.Org, projectName, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				} else {
					err := listProjectMembers(cmd, ch, ch.Org, projectName, pageToken, pageSize)
					if err != nil {
						return err
					}

					err = listProjectInvites(cmd, ch, ch.Org, projectName, pageToken, pageSize)
					if err != nil {
						return err
					}
				}

				// TODO: user groups
			} else {
				if strings.HasPrefix(pageToken, "usr") {
					err := listOrgMembers(cmd, ch, ch.Org, strings.TrimPrefix(pageToken, "usr"), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, "inv") {
					err := listOrgInvites(cmd, ch, ch.Org, strings.TrimPrefix(pageToken, "inv"), pageSize)
					if err != nil {
						return err
					}
				} else {
					err := listOrgMembers(cmd, ch, ch.Org, pageToken, pageSize)
					if err != nil {
						return err
					}

					err = listOrgInvites(cmd, ch, ch.Org, pageToken, pageSize)
					if err != nil {
						return err
					}
				}

				// TODO: user groups
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")
	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of users to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}

func listProjectMembers(cmd *cobra.Command, ch *cmdutil.Helper, org, project, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListProjectMembers(context.Background(), &adminv1.ListProjectMembersRequest{
		Organization: org,
		Project:      project,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintMembers(members.Members)

	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
	}

	return nil
}

func listProjectInvites(cmd *cobra.Command, ch *cmdutil.Helper, org, project, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	invites, err := client.ListProjectInvites(context.Background(), &adminv1.ListProjectInvitesRequest{
		Organization: org,
		Project:      project,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}
	// If page token is empty, user is running the command first time and we print separator
	if len(invites.Invites) > 0 && pageToken == "" {
		cmd.Println()
	}
	ch.PrintfSuccess("Pending user invites\n")
	ch.PrintInvites(invites.Invites)

	if invites.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: inv%s\n", invites.NextPageToken)
	}

	return nil
}

func listOrgMembers(cmd *cobra.Command, ch *cmdutil.Helper, org, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListOrganizationMembers(context.Background(), &adminv1.ListOrganizationMembersRequest{
		Organization: org,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintMembers(members.Members)

	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: usr%s\n", members.NextPageToken)
	}
	return nil
}

func listOrgInvites(cmd *cobra.Command, ch *cmdutil.Helper, org, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	invites, err := client.ListOrganizationInvites(context.Background(), &adminv1.ListOrganizationInvitesRequest{
		Organization: org,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return err
	}
	// If page token is empty, user is running the command first time and we print separator
	if len(invites.Invites) > 0 && pageToken == "" {
		cmd.Println()
	}
	ch.PrintfSuccess("Pending user invites\n")
	ch.PrintInvites(invites.Invites)

	if invites.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: inv%s\n", invites.NextPageToken)
	}

	return nil
}
