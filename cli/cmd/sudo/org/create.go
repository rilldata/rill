package org

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, description, plan string

	createCmd := &cobra.Command{
		Use:   "create [<org-name>]",
		Short: "Create organization for internal use with unlimited quotas",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if len(args) == 0 && ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "name")
				if err != nil {
					return err
				}
			}

			res, err := client.CreateOrganization(cmd.Context(), &adminv1.CreateOrganizationRequest{
				Name:                 name,
				Description:          description,
				SuperuserForceAccess: true,
			})
			if err != nil {
				if !isNameExistsErr(err) {
					return err
				}

				fmt.Printf("Org name %q already exists\n", name)
				return nil
			}

			// Switching to the created org
			org := res.Organization
			err = dotrill.SetDefaultOrg(org.Name)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Created organization\n")
			ch.PrintOrgs([]*adminv1.Organization{org}, "")

			ch.PrintfSuccess("\nSubscribing org to the plan %q, this might take few seconds\n", plan)

			// org billing init is async, so we need to wait for the billing to be initialized
			timeout := time.After(10 * time.Second)
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-cmd.Context().Done():
					return cmd.Context().Err()
				case <-timeout:
					ch.PrintfError("\nTimed out waiting for billing to be initialized\n")
					ch.PrintfWarn(fmt.Sprintf("\nRun 'rill billing subscription edit --plan %s --force' to subscribe to the plan manually\n", plan))
					return err
				case <-ticker.C:
					res, err := client.UpdateBillingSubscription(cmd.Context(), &adminv1.UpdateBillingSubscriptionRequest{
						Organization:         org.Name,
						PlanName:             plan,
						SuperuserForceAccess: true,
					})
					if err == nil {
						ch.PrintfSuccess("\nSubscribed organization to plan %q\n", plan)
						ch.PrintSubscriptions([]*adminv1.Subscription{res.Subscription})
						return nil
					} else if strings.Contains(err.Error(), "billing not yet initialized") {
						fmt.Println("Waiting for billing to be initialized...")
					} else {
						ch.PrintfError(fmt.Sprintf("\nFailed to subscribe organization to plan %q with error %v\n", plan, err))
						ch.PrintfWarn("\nDeleting organization %q\n", org.Name)
						_, err = client.DeleteOrganization(cmd.Context(), &adminv1.DeleteOrganizationRequest{
							Name: org.Name,
						})
						if err != nil {
							ch.PrintfError("\nFailed to delete organization %q with error %v\n", org.Name, err)
						}
						return err
					}
				}
			}
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&name, "name", "", "Organization Name")
	createCmd.Flags().StringVar(&description, "description", "", "Description")
	createCmd.Flags().StringVar(&plan, "plan", "superuser", "Plan to subscribe to")
	return createCmd
}

func isNameExistsErr(err error) bool {
	if strings.Contains(err.Error(), "already exists") {
		return true
	}
	if strings.Contains(err.Error(), "violates unique constraint") {
		return true
	}
	return false
}
