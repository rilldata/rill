package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var group string
	var email string
	var role string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add",
		RunE: func(cmd *cobra.Command, args []string) error {
			if group == "" {
				err := cmdutil.SelectPromptIfEmpty(&role, "Select role", userRoles, "")
				if err != nil {
					return err
				}
			}

			err := cmdutil.StringPromptIfEmpty(&email, "Enter email")
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if group != "" {
				res, err := client.AddUsergroupMemberUser(cmd.Context(), &adminv1.AddUsergroupMemberUserRequest{
					Organization:     ch.Org,
					Usergroup:        group,
					Email:            email,
					AddRillUserToOrg: false,
				})
				if err != nil {
					if s, ok := status.FromError(err); ok && s.Code() == codes.FailedPrecondition && strings.Contains(s.Message(), "user is not a member of the org") {
						// ask if user should be added to the organization
						addToOrg, err := cmdutil.ConfirmPrompt(fmt.Sprintf("This user exists in Rill but not part of the %q organization. Do you want to add the user to the organization as viewer?", ch.Org), "", false)
						if err != nil {
							return err
						}

						if addToOrg {
							_, err = client.AddUsergroupMemberUser(cmd.Context(), &adminv1.AddUsergroupMemberUserRequest{
								Organization:     ch.Org,
								Usergroup:        group,
								Email:            email,
								AddRillUserToOrg: true,
							})
							if err != nil {
								return err
							}
							ch.PrintfSuccess("User %q added to the user group %q and organization %q\n", email, group, ch.Org)
							return nil
						}
						ch.PrintfWarn("Aborted, user needs to be part of the organization to be added to the user group\n")
						return nil
					}
					return err
				}

				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join user group \"%s\"\n", email, group)
				} else {
					ch.PrintfSuccess("User %q added to the user group %q\n", email, group)
				}
			} else if projectName != "" {
				res, err := client.AddProjectMemberUser(cmd.Context(), &adminv1.AddProjectMemberUserRequest{
					Organization: ch.Org,
					Project:      projectName,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				} else {
					ch.PrintfSuccess("User %q added to the project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				}
			} else {
				res, err := client.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
					Organization: ch.Org,
					Email:        email,
					Role:         role,
				})
				if err != nil {
					return err
				}

				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q\n", email, ch.Org, role)
				} else {
					ch.PrintfSuccess("User %q added to the organization %q as %q\n", email, ch.Org, role)
				}
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&group, "group", "", "User group")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user (options: %s)", strings.Join(userRoles, ", ")))

	return addCmd
}
