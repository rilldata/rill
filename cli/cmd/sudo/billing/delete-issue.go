package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteIssueCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, issueType string
	errors := []string{"on-trial", "trial-ended", "no-payment-method", "no-billable-address", "payment-failed", "subscription-cancelled", "never-subscribed"}
	setCmd := &cobra.Command{
		Use:   "delete-issue",
		Short: "Delete billing issue of a type for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if org == "" {
				return fmt.Errorf("please set --org")
			}

			if issueType == "" {
				issueType, err = cmdutil.SelectPrompt("Select issue type to delete", errors, "")
				if err != nil {
					return err
				}
			}

			var t adminv1.BillingIssueType
			switch issueType {
			case "on-trial":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL
			case "trial-ended":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED
			case "no-payment-method":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD
			case "no-billable-address":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS
			case "payment-failed":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED
			case "subscription-cancelled":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED
			case "never-subscribed":
				t = adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NEVER_SUBSCRIBED
			default:
				return fmt.Errorf("invalid error type %q", issueType)
			}

			_, err = client.SudoDeleteOrganizationBillingIssue(ctx, &adminv1.SudoDeleteOrganizationBillingIssueRequest{
				Org:  org,
				Type: t,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Billing issue of type %q deleted for organization %q\n", issueType, org)

			return nil
		},
	}

	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&issueType, "type", "", "Billing Issue Type")
	return setCmd
}
