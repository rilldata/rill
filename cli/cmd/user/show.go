package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string

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

			return nil
		},
	}

	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	showCmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")

	return showCmd
}
