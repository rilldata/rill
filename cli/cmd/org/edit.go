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
			if !cmd.Flags().Changed("org") && len(args) == 0 && ch.Interactive {
				orgNames, err := OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				orgName, err = cmdutil.SelectPrompt("Select org to edit", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: orgName})
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
				Name: org.Name,
			}

			if cmd.Flags().Changed("display-name") {
				req.DisplayName = &displayName
			} else if ch.Interactive {
				ok, err := cmdutil.ConfirmPrompt("Do you want to update the display name", "", false)
				if err != nil {
					return err
				}
				if ok {
					displayName, err = cmdutil.InputPrompt("Enter the display name", org.DisplayName)
					if err != nil {
						return err
					}
					req.DisplayName = &displayName
				}
			}

			if cmd.Flags().Changed("description") {
				req.Description = &description
			} else if ch.Interactive {
				ok, err := cmdutil.ConfirmPrompt("Do you want to update the description", "", false)
				if err != nil {
					return err
				}
				if ok {
					description, err = cmdutil.InputPrompt("Enter the description", org.Description)
					if err != nil {
						return err
					}
					req.Description = &description
				}
			}

			if cmd.Flags().Changed("default-project-role") {
				if defaultProjectRole == "none" {
					defaultProjectRole = ""
				}
				req.DefaultProjectRole = &defaultProjectRole
			}

			if cmd.Flags().Changed("billing-email") {
				req.BillingEmail = &billingEmail
			} else if ch.Interactive {
				ok, err := cmdutil.ConfirmPrompt("Do you want to update the billing email", "", false)
				if err != nil {
					return err
				}
				if ok {
					billingEmail, err = cmdutil.InputPrompt("Enter the billing email", org.BillingEmail)
					if err != nil {
						return err
					}
					req.BillingEmail = &billingEmail
				}
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
