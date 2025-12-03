package usergroup

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetResourcesCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var groupName string
	var role string
	var explores []string
	var canvases []string
	var restrict bool

	cmd := &cobra.Command{
		Use:   "set-resources",
		Short: "Set a user group's project resources and restriction flag (overwrites existing list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectName == "" {
				return fmt.Errorf("--project is required")
			}

			if groupName == "" && ch.Interactive {
				if err := cmdutil.StringPromptIfEmpty(&groupName, "Enter user group name"); err != nil {
					return err
				}
			}
			if groupName == "" {
				return fmt.Errorf("--group is required")
			}

			resources, err := cmdutil.ParseResourceStrings(explores, canvases)
			if err != nil {
				return err
			}
			if len(resources) == 0 && !cmd.Flags().Changed("restrict-resources") {
				// error out if no resources and restrict not explicitly set
				return fmt.Errorf("either resources must be provided or --restrict-resources must be set to true or false to enforce or clear restrictions")
			}

			if len(resources) > 0 {
				restrict = true
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.SetProjectMemberUsergroupRole(cmd.Context(), &adminv1.SetProjectMemberUsergroupRoleRequest{
				Org:               ch.Org,
				Project:           projectName,
				Usergroup:         groupName,
				Resources:         resources,
				RestrictResources: &restrict,
			})
			if err != nil {
				return err
			}

			resourceList := cmdutil.FormatResourceNames(resources)
			status := "cleared"
			if restrict && len(resources) > 0 {
				status = fmt.Sprintf("set to %s", resourceList)
			} else if restrict {
				status = "restricted with no resources"
			}
			ch.PrintfSuccess("Updated resources for user group %q in project \"%s/%s\" (role %s, resources %s)\n", groupName, ch.Org, projectName, role, status)
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&projectName, "project", "", "Project (required)")
	cmd.Flags().StringVar(&groupName, "group", "", "User group (required)")
	cmd.Flags().StringArrayVar(&explores, "explore", nil, "Explore resource to restrict to (repeat for multiple)")
	cmd.Flags().StringArrayVar(&canvases, "canvas", nil, "Canvas resource to restrict to (repeat for multiple)")
	cmd.Flags().BoolVar(&restrict, "restrict-resources", false, "Whether to restrict the group to the provided resources (defaults to true when resources are provided)")

	return cmd
}
