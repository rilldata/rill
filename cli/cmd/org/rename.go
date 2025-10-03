package org

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName, displayName string

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

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: name})
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
					ch.PrintfError("org %q doesn't exist, run 'rill org list' to see available orgs", name)
					return nil
				}
				return err
			}
			org := resp.Organization

			// Build update request
			req := &adminv1.UpdateOrganizationRequest{Org: org.Name}

			var flagSet bool
			if cmd.Flags().Changed("new-name") {
				flagSet = true
				req.NewName = &newName
			}

			if cmd.Flags().Changed("display-name") {
				flagSet = true
				req.DisplayName = &displayName
			}

			if !flagSet {
				ch.PrintfError("no changes requested please specify --new-name or --display-name\n")
				return nil
			}

			if req.NewName != nil {
				ch.PrintfWarn("Changing the name will invalidate dashboard URLs.\n")
				ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			// Update org
			updatedOrg, err := client.UpdateOrganization(ctx, req)
			if err != nil {
				return err
			}

			if req.NewName != nil {
				ch.Printf("Updated name %q to %q\n", name, updatedOrg.Organization.Name)
				if org.DisplayName != "" && req.DisplayName == nil {
					ch.PrintfWarn("You updated the org's unique name, but not its display name; use --display-name to update the display name\n")
				}
			}

			if req.DisplayName != nil {
				ch.Printf("Updated display name: %s to %s\n", org.DisplayName, updatedOrg.Organization.DisplayName)
			}

			ch.PrintOrgs([]*adminv1.Organization{updatedOrg.Organization}, "")

			// Update default org if name changed
			if req.NewName != nil {
				err = ch.DotRill.SetDefaultOrg(*req.NewName)
				if err != nil {
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

	return renameCmd
}
