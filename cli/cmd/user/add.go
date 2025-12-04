package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string
	var projectName string
	var group string
	var role string // NOTE: Overloaded to mean org role or project role based on whether --project is specified

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add user to a project, organization or group",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Handle empty email
			if email == "" {
				if !ch.Interactive {
					return fmt.Errorf("--email is required when not running interactively")
				}
				err = cmdutil.StringPromptIfEmpty(&email, "Enter email")
				if err != nil {
					return err
				}
			}

			// Handle adding the user to the org.
			// We do this only if a more specific target (group or project) is not specified.
			if projectName == "" && group == "" {
				// Handle empty role
				if role == "" {
					if !ch.Interactive {
						return fmt.Errorf("--role is required when not running interactively")
					}
					err := cmdutil.SelectPromptIfEmpty(&role, "Select role", orgRoles, orgRoles[len(orgRoles)-1])
					if err != nil {
						return err
					}
				}

				// Add to org
				res, err := client.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
					Org:   ch.Org,
					Email: email,
					Role:  role,
				})
				if err != nil {
					return err
				}

				// Print status and exit
				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q\n", email, ch.Org, role)
				} else {
					ch.PrintfSuccess("User %q added to the organization %q as %q\n", email, ch.Org, role)
				}
				return nil
			}

			// Handle adding the user to a project.
			if projectName != "" {
				// Handle empty role
				if role == "" {
					if !ch.Interactive {
						return fmt.Errorf("--role is required when not running interactively")
					}
					err := cmdutil.SelectPromptIfEmpty(&role, "Select role", projectRoles, projectRoles[len(projectRoles)-1])
					if err != nil {
						return err
					}
				}

				// Add to project
				res, err := client.AddProjectMemberUser(cmd.Context(), &adminv1.AddProjectMemberUserRequest{
					Org:     ch.Org,
					Project: projectName,
					Email:   email,
					Role:    role,
				})
				if err != nil {
					// We don't need to handle org membership errors since AddProjectMemberUser automatically invites the user to the org with role guest if needed.
					return err
				}

				// Print status
				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				} else {
					ch.PrintfSuccess("User %q added to the project \"%s/%s\" as %q\n", email, ch.Org, projectName, role)
				}

				// NOTE: Not exiting here since we may also need to add to a group.
			}

			// Handle adding the user to a group.
			if group != "" {
				_, err = client.AddUsergroupMemberUser(cmd.Context(), &adminv1.AddUsergroupMemberUserRequest{
					Org:       ch.Org,
					Usergroup: group,
					Email:     email,
				})
				if err != nil {
					// If the user is not in the organization, we'll try to interactively add them.
					if !strings.Contains(err.Error(), "user is not a member of the org") {
						return err
					}
					if !ch.Interactive {
						return err
					}
					ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("The user must be a member of %q to join one of its groups. Do you want to invite the user to join %q?", ch.Org, ch.Org), "", false)
					if err != nil {
						return err
					}
					if !ok {
						return fmt.Errorf("aborted: user needs to be part of the organization to be added to the user group")
					}

					// Find the org role to use.
					orgRole := role
					if projectName != "" {
						// When --project is specified, --role refers to the project role, not the org role.
						orgRole = ""
					}
					if orgRole == "" {
						err := cmdutil.SelectPromptIfEmpty(&orgRole, "Select organization role", orgRoles, orgRoles[len(orgRoles)-1])
						if err != nil {
							return err
						}
					}

					// Add the user to the organization
					res, err := client.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
						Org:   ch.Org,
						Email: email,
						Role:  orgRole,
					})
					if err != nil {
						return err
					}

					// Print status
					if res.PendingSignup {
						ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q\n", email, ch.Org, orgRole)
					} else {
						ch.PrintfSuccess("User %q added to the organization %q as %q\n", email, ch.Org, orgRole)
					}

					// User is now in the org, retry adding to the group
					_, err = client.AddUsergroupMemberUser(cmd.Context(), &adminv1.AddUsergroupMemberUserRequest{
						Org:       ch.Org,
						Usergroup: group,
						Email:     email,
					})
					if err != nil {
						return err
					}
				}

				// Print status
				ch.PrintfSuccess("User %q added to the user group %q\n", email, group)
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringVar(&group, "group", "", "User group")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user (options: %s)", strings.Join(orgRoles, ", ")))

	return addCmd
}
