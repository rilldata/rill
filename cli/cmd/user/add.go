package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string
	var projectName string
	var groups []string
	var role string // NOTE: Overloaded to mean org role or project role based on whether --project is specified
	var explores []string
	var canvases []string
	var restrictResources bool
	var attributes map[string]string
	var attributesJSON string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add user to a project, organization or group",
		Long: `Add user to a project, organization or group.

When --group is given together with --project, or with --role and no --project, the user is added to (or invited to)
the project or organization and its user groups in a single step. A user who has not signed up yet joins the groups when
they accept the invitation. When only --group is given, the user is added to the groups directly.`,
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
				email, err = cmdutil.InputPrompt("Enter email", "")
				if err != nil {
					return err
				}
			}

			if projectName == "" && (len(explores) > 0 || len(canvases) > 0 || restrictResources) {
				return fmt.Errorf("resource restrictions can only be set when adding a user to a project")
			}

			// Parse custom attributes if provided. They apply to the user's org membership.
			var attrsPB *structpb.Struct
			if len(attributes) > 0 || attributesJSON != "" {
				attrs, err := parseAttributes(attributes, attributesJSON)
				if err != nil {
					return err
				}
				attrsPB, err = structpb.NewStruct(attrs)
				if err != nil {
					return fmt.Errorf("failed to parse attributes: %w", err)
				}
			}

			// Handle adding the user to the org (and optionally its groups).
			// We do this if no project is specified, unless only groups are specified (handled last).
			if projectName == "" && (len(groups) == 0 || role != "") {
				// Handle empty role
				if role == "" {
					if !ch.Interactive {
						return fmt.Errorf("--role is required when not running interactively")
					}
					role, err = cmdutil.SelectPrompt("Select role", orgRoles, orgRoles[len(orgRoles)-1])
					if err != nil {
						return err
					}
				}

				// Add to org
				res, err := client.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
					Org:        ch.Org,
					Email:      email,
					Role:       role,
					Attributes: attrsPB,
					Usergroups: groups,
				})
				if err != nil {
					return err
				}

				// Print status and exit
				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q%s\n", email, ch.Org, role, groupsSuffix(groups, true))
				} else {
					ch.PrintfSuccess("User %q added to the organization %q as %q%s\n", email, ch.Org, role, groupsSuffix(groups, false))
				}
				return nil
			}

			// Handle adding the user to a project (and optionally its org's groups).
			if projectName != "" {
				resources, err := cmdutil.ParseResourceStrings(explores, canvases)
				if err != nil {
					return err
				}
				if len(resources) > 0 {
					restrictResources = true
				}

				// Handle empty role
				if role == "" {
					if !ch.Interactive {
						return fmt.Errorf("--role is required when not running interactively")
					}
					role, err = cmdutil.SelectPrompt("Select role", projectRoles, projectRoles[len(projectRoles)-1])
					if err != nil {
						return err
					}
				}

				// Add to project
				res, err := client.AddProjectMemberUser(cmd.Context(), &adminv1.AddProjectMemberUserRequest{
					Org:               ch.Org,
					Project:           projectName,
					Email:             email,
					Role:              role,
					Resources:         resources,
					RestrictResources: &restrictResources,
					Attributes:        attrsPB,
					Usergroups:        groups,
				})
				if err != nil {
					// We don't need to handle org membership errors since AddProjectMemberUser automatically invites the user to the org with role guest if needed.
					return err
				}

				// Print status and exit
				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join project \"%s/%s\" as %q%s\n", email, ch.Org, projectName, role, groupsSuffix(groups, true))
				} else {
					ch.PrintfSuccess("User %q added to the project \"%s/%s\" as %q%s\n", email, ch.Org, projectName, role, groupsSuffix(groups, false))
				}
				return nil
			}

			// Handle adding the user to groups only.
			// This works for org members and for users with a pending org invite (they join the groups on acceptance).
			for _, group := range groups {
				_, err = client.AddUsergroupMemberUser(cmd.Context(), &adminv1.AddUsergroupMemberUserRequest{
					Org:       ch.Org,
					Usergroup: group,
					Email:     email,
				})
				if err == nil {
					ch.PrintfSuccess("User %q added to the user group %q\n", email, group)
					continue
				}

				// If the user is not in the organization, we'll try to interactively invite them to the org and all the groups in one go.
				if !strings.Contains(err.Error(), "user is not a member of the org") {
					return err
				}
				if !ch.Interactive {
					return err
				}
				if err := cmdutil.ConfirmPrompt(fmt.Sprintf("The user must be a member of %q to join one of its groups. Do you want to invite the user to join %q?", ch.Org, ch.Org), false); err != nil {
					return err
				}

				orgRole, err := cmdutil.SelectPrompt("Select organization role", orgRoles, orgRoles[len(orgRoles)-1])
				if err != nil {
					return err
				}

				// Add the user to the organization and its groups
				res, err := client.AddOrganizationMemberUser(cmd.Context(), &adminv1.AddOrganizationMemberUserRequest{
					Org:        ch.Org,
					Email:      email,
					Role:       orgRole,
					Attributes: attrsPB,
					Usergroups: groups,
				})
				if err != nil {
					return err
				}

				// Print status and exit
				if res.PendingSignup {
					ch.PrintfSuccess("Invitation sent to %q to join organization %q as %q%s\n", email, ch.Org, orgRole, groupsSuffix(groups, true))
				} else {
					ch.PrintfSuccess("User %q added to the organization %q as %q%s\n", email, ch.Org, orgRole, groupsSuffix(groups, false))
				}
				return nil
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	addCmd.Flags().StringVar(&projectName, "project", "", "Project")
	addCmd.Flags().StringArrayVar(&groups, "group", nil, "User group to add the user to (repeat for multiple)")
	addCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	addCmd.Flags().StringVar(&role, "role", "", fmt.Sprintf("Role of the user (options: %s)", strings.Join(orgRoles, ", ")))
	addCmd.Flags().StringArrayVar(&explores, "explore", nil, "Explore resource to restrict to (repeat for multiple)")
	addCmd.Flags().StringArrayVar(&canvases, "canvas", nil, "Canvas resource to restrict to (repeat for multiple)")
	addCmd.Flags().BoolVar(&restrictResources, "restrict-resources", false, "Restrict the user to the provided resources (defaults to true when resources are provided)")
	addCmd.Flags().StringToStringVar(&attributes, "attribute", nil, "Custom attributes in key=value format (--attribute app=foo --attribute dept=bar)")
	addCmd.Flags().StringVar(&attributesJSON, "json", "", "Custom attributes as JSON object (--json '{\"app\":\"foo\",\"dept\":\"bar\"}')")

	return addCmd
}

// groupsSuffix describes the user groups in a success message.
// It is empty when no groups were given.
func groupsSuffix(groups []string, pending bool) string {
	if len(groups) == 0 {
		return ""
	}
	noun := "user group"
	if len(groups) > 1 {
		noun = "user groups"
	}
	if pending {
		return fmt.Sprintf(" (will be added to %s %s on acceptance)", noun, strings.Join(groups, ", "))
	}
	return fmt.Sprintf(" and to %s %s", noun, strings.Join(groups, ", "))
}
