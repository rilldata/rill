package usergroup

import (
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
		Short: "List groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectName != "" {
				err := listProjectMemberUsergroups(cmd, ch, ch.Org, projectName, pageToken, pageSize)
				if err != nil {
					return err
				}
			} else {
				err := listOrgMemberUsergroups(cmd, ch, ch.Org, pageToken, pageSize)
				if err != nil {
					return err
				}

				ch.Printf("\nShowing organization user groups only. Use the --project flag to list user groups of a specific project.\n")
			}

			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	listCmd.Flags().StringVar(&projectName, "project", "", "Project")
	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of user groups to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}

func listProjectMemberUsergroups(cmd *cobra.Command, ch *cmdutil.Helper, org, project, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListProjectMemberUsergroups(cmd.Context(), &adminv1.ListProjectMemberUsergroupsRequest{
		Org:       org,
		Project:   project,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintMemberUsergroups(members.Members)

	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: %s\n", members.NextPageToken)
	}

	return nil
}

func listOrgMemberUsergroups(cmd *cobra.Command, ch *cmdutil.Helper, org, pageToken string, pageSize uint32) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	members, err := client.ListOrganizationMemberUsergroups(cmd.Context(), &adminv1.ListOrganizationMemberUsergroupsRequest{
		Org:       org,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return err
	}

	ch.PrintMemberUsergroups(members.Members)

	if members.NextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: %s\n", members.NextPageToken)
	}
	return nil
}
