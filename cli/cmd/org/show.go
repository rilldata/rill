package org

import (
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show [<org-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Show org details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				ch.Org = args[0]
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetOrganization(cmd.Context(), &adminv1.GetOrganizationRequest{
				Name: ch.Org,
			})
			if err != nil {
				return err
			}
			org := res.Organization

			ch.Printf("Id: %s\n", org.Id)
			ch.Printf("Name: %s\n", org.Name)
			ch.Printf("Display Name: %s\n", org.DisplayName)
			ch.Printf("Description: %s\n", org.Description)
			ch.Printf("Custom Logo: %s\n", org.LogoUrl)
			ch.Printf("Custom Favicon: %s\n", org.FaviconUrl)
			ch.Printf("Custom Domain: %s\n", org.CustomDomain)
			ch.Printf("Billing Email: %s\n", org.BillingEmail)
			ch.Printf("Created On: %s\n", org.CreatedOn.AsTime().Format(time.RFC3339Nano))
			ch.Printf("Updated On: %s\n", org.UpdatedOn.AsTime().Format(time.RFC3339Nano))

			return nil
		},
	}

	showCmd.Flags().SortFlags = false
	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization name")

	return showCmd
}
