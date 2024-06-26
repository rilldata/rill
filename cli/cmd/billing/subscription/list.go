package subscription

import (
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("org") && ch.Interactive {
				orgNames, err := org.OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				orgName, err = cmdutil.SelectPrompt("Select org to list subscription", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetOrganizationBillingSubscription(cmd.Context(), &adminv1.GetOrganizationBillingSubscriptionRequest{
				OrgName: orgName,
			})
			if err != nil {
				return err
			}

			if resp.Subscription == nil {
				ch.PrintfWarn("No subscription found for organization %q.\n", orgName)
				return nil
			}

			ch.PrintSubscriptions([]*adminv1.Subscription{resp.Subscription})
			return nil
		},
	}
	listCmd.Flags().SortFlags = false
	listCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization Name")
	return listCmd
}
