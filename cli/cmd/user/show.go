package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string
	var project string

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show detailed information about a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if email == "" {
				return fmt.Errorf("email is required")
			}

			member, err := client.GetOrganizationMemberUser(cmd.Context(), &adminv1.GetOrganizationMemberUserRequest{
				Org:   ch.Org,
				Email: email,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User: %s (%s)\n", member.Member.UserName, member.Member.UserEmail)
			ch.PrintfSuccess("Role: %s\n", member.Member.RoleName)
			ch.PrintfSuccess("Projects count: %d\n", member.Member.ProjectsCount)
			ch.PrintfSuccess("Usergroups count: %d\n", member.Member.UsergroupsCount)

			attrs := member.Member.Attributes.AsMap()
			if len(attrs) > 0 {
				ch.PrintfSuccess("Attributes:\n")
				for key, value := range attrs {
					ch.PrintfSuccess("  %s: %v\n", key, value)
				}
			} else {
				ch.PrintfSuccess("No custom attributes set\n")
			}

			if project == "" {
				return nil
			}
			projMember, err := client.GetProjectMemberUser(cmd.Context(), &adminv1.GetProjectMemberUserRequest{
				Org:     ch.Org,
				Project: project,
				Email:   email,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				return err
			}
			if projMember != nil {
				ch.PrintfSuccess("Project Member Info:\n")
				ch.PrintProjectMemberUsers([]*adminv1.ProjectMemberUser{projMember.Member})
			} else {
				cmd.Printf("No membership found for user %q in project %q\n", email, project)
			}

			groupResp, err := client.ListUsergroupsForProjectAndUser(cmd.Context(), &adminv1.ListUsergroupsForProjectAndUserRequest{
				Org:     ch.Org,
				Project: project,
				Email:   email,
			})
			if err != nil {
				return err
			}
			if len(groupResp.Usergroups) > 0 {
				ch.PrintfSuccess("\nThe user has project access through the following user groups:\n")
				ch.PrintMemberUsergroups(groupResp.Usergroups)
			}

			return nil
		},
	}

	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	showCmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	showCmd.Flags().StringVar(&project, "project", "", "Project name to include project membership details (optional)")

	return showCmd
}
