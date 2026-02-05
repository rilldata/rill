package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName, displayName, description, defaultProjectRole, billingEmail string

	editCmd := &cobra.Command{
		Use:   "edit [<org-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit organization details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				orgName = args[0]
			}
			if orgName == "" {
				return fmt.Errorf("must specify an organization name")
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: orgName})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					if st.Code() != codes.NotFound {
						return err
					}
				}
				fmt.Printf("Org name %q doesn't exist, please run `rill org list` to list available orgs\n", orgName)
				return nil
			}

			org := resp.Organization
			req := &adminv1.UpdateOrganizationRequest{
				Org: org.Name,
			}

			var flagSet bool
			if cmd.Flags().Changed("display-name") {
				flagSet = true
				req.DisplayName = &displayName
			}

			if cmd.Flags().Changed("description") {
				flagSet = true
				req.Description = &description
			}

			if cmd.Flags().Changed("default-project-role") {
				if defaultProjectRole == "none" {
					defaultProjectRole = ""
				}
				req.DefaultProjectRole = &defaultProjectRole
			}

			if cmd.Flags().Changed("billing-email") {
				flagSet = true
				req.BillingEmail = &billingEmail
			}

			if !flagSet {
				return fmt.Errorf("at least one flag must be set")
			}

			updatedOrg, err := client.UpdateOrganization(ctx, req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated organization\n")
			ch.PrintOrgs([]*adminv1.Organization{updatedOrg.Organization}, "")

			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization name")
	editCmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")
	editCmd.Flags().StringVar(&defaultProjectRole, "default-project-role", "", "Default role for members on new projects (options: admin, editor, viewer, none)")
	editCmd.Flags().StringVar(&billingEmail, "billing-email", "", "Billing email")

	return editCmd
}
