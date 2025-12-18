package org

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetInternalPlanCmd(ch *cmdutil.Helper) *cobra.Command {
	var plan string

	setInternalPlanCmd := &cobra.Command{
		Use:   "set-internal-plan [<org-name>]",
		Short: "Set internal plan with unlimited quotas",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				return fmt.Errorf("org name is required")
			}

			name := args[0]

			ch.PrintfSuccess("Subscribing org to the plan %q, this might take few seconds\n", plan)

			// org billing init is async, so we need to wait for the billing to be initialized
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-cmd.Context().Done():
					return cmd.Context().Err()
				case <-timeout:
					ch.PrintfError("\nTimed out waiting for billing to be initialized\n")
					ch.PrintfWarn("\nRun 'rill billing subscription edit --plan %s --force' to subscribe to the plan manually\n", plan)
					return err
				case <-ticker.C:
					res, err := client.UpdateBillingSubscription(cmd.Context(), &adminv1.UpdateBillingSubscriptionRequest{
						Org:                  name,
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
						return err
					}
				}
			}
		},
	}

	setInternalPlanCmd.Flags().StringVar(&plan, "plan", "superuser", "Plan to subscribe to")
	return setInternalPlanCmd
}
