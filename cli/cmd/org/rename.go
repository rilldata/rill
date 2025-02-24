package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName, displayName string
	var force bool

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("new-name") && !cmd.Flags().Changed("display-name") {
				return fmt.Errorf("at least one of --new-name or --display-name must be provided")
			}

			if !cmd.Flags().Changed("org") && len(args) == 0 && ch.Interactive {
				orgNames, err := OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				name, err = cmdutil.SelectPrompt("Select org to edit", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
					return fmt.Errorf("org %q doesn't exist, run 'rill org list' to see available orgs", name)
				}
				return err
			}
			org := resp.Organization

			// Require at least one change
			if !cmd.Flags().Changed("new-name") && !cmd.Flags().Changed("display-name") {
				return fmt.Errorf("at least one of --new-name or --display-name must be provided")
			}

			// Build update request
			req := &adminv1.UpdateOrganizationRequest{
				Name: org.Name,
			}

			if cmd.Flags().Changed("new-name") {
				ch.PrintfWarn("\nWarn: Changing org name will invalidate dashboard URLs.\n")
				if !force {
					ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
					if err != nil {
						return err
					}
					if !ok {
						return fmt.Errorf("operation cancelled")
					}
				}
				req.NewName = &newName
			}

			if cmd.Flags().Changed("display-name") {
				req.DisplayName = &displayName
			}

			// Update org
			updatedOrg, err := client.UpdateOrganization(ctx, req)
			if err != nil {
				return err
			}

			// Print results
			ch.PrintfSuccess("Updated organization\n")
			if req.NewName != nil {
				ch.Printf("Updated name: %s to %s\n", name, updatedOrg.Organization.Name)
			}
			if req.DisplayName != nil {
				ch.Printf("Updated display name: %s to %s\n", org.DisplayName, updatedOrg.Organization.DisplayName)
			}

			ch.PrintOrgs([]*adminv1.Organization{updatedOrg.Organization}, "")

			// Update default org if name changed
			if req.NewName != nil {
				if err := dotrill.SetDefaultOrg(*req.NewName); err != nil {
					return err
				}
			}

			return nil
		},
	}

	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "org", ch.Org, "Current org name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New org name")
	renameCmd.Flags().StringVar(&displayName, "display-name", "", "New display name")
	renameCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")

	return renameCmd
}
