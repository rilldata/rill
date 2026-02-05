package user

import (
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

const (
	userTokenPrefix   = "usr" // User token prefix
	inviteTokenPrefix = "inv" // Invite token prefix
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var groupName string
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupName != "" {
				err := listUsergroupMembers(cmd, ch, ch.Org, groupName, pageToken, pageSize)
				if err != nil {
					return err
				}
			} else if projectName != "" {
				if strings.HasPrefix(pageToken, userTokenPrefix) {
					err := listProjectMembers(cmd, ch, ch.Org, projectName, strings.TrimPrefix(pageToken, userTokenPrefix), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, inviteTokenPrefix) {
					err := listProjectInvites(cmd, ch, ch.Org, projectName, strings.TrimPrefix(pageToken, inviteTokenPrefix), pageSize)
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
			} else {
				if strings.HasPrefix(pageToken, userTokenPrefix) {
					err := listOrgMembers(cmd, ch, ch.Org, strings.TrimPrefix(pageToken, userTokenPrefix), pageSize)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(pageToken, inviteTokenPrefix) {
					err := listOrgInvites(cmd, ch, ch.Org, strings.TrimPrefix(pageToken, inviteTokenPrefix), pageSize)
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
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")
	listCmd.Flags().StringVar(&groupName, "group", "", "User group")
	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of users to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}

func listUsergroupMembers(cmd *cobra.Command, ch *cmdutil.Helper, org, group, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListUsergroupMemberUsers(cmd.Context(), &adminv1.ListUsergroupMemberUsersRequest{
		Org:       org,
		Usergroup: group,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintUsergroupMemberUsers(members.Members)

	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: %s\n", members.NextPageToken)
	}

	return nil
}

func listProjectMembers(cmd *cobra.Command, ch *cmdutil.Helper, org, project, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListProjectMemberUsers(cmd.Context(), &adminv1.ListProjectMemberUsersRequest{
		Org:       org,
		Project:   project,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintProjectMemberUsers(members.Members)

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

	invites, err := client.ListProjectInvites(cmd.Context(), &adminv1.ListProjectInvitesRequest{
		Org:       org,
		Project:   project,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	if len(invites.Invites) == 0 {
		return nil
	}

	// If page token is empty, user is running the command first time and we print separator
	if pageToken == "" {
		cmd.Println()
	}

	ch.PrintProjectInvites(invites.Invites)

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

	members, err := client.ListOrganizationMemberUsers(cmd.Context(), &adminv1.ListOrganizationMemberUsersRequest{
		Org:       org,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintOrganizationMemberUsers(members.Members)

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

	invites, err := client.ListOrganizationInvites(cmd.Context(), &adminv1.ListOrganizationInvitesRequest{
		Org:       org,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	if len(invites.Invites) == 0 {
		return nil
	}

	// If page token is empty, user is running the command first time and we print separator
	if pageToken == "" {
		cmd.Println()
	}

	ch.PrintOrganizationInvites(invites.Invites)

	if invites.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: inv%s\n", invites.NextPageToken)
	}

	return nil
}
