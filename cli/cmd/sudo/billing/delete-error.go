package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteErrorCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, errorType string
	errors := []string{"no-payment-method", "trial-ended", "payment-failed"}
	setCmd := &cobra.Command{
		Use:   "delete-error",
		Short: "Delete billing error of a type for an organization",
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

			if errorType == "" {
				errorType, err = cmdutil.SelectPrompt("Select error type to delete", errors, "")
				if err != nil {
					return err
				}
			}

			var t adminv1.BillingErrorType
			switch errorType {
			case "no-payment-method":
				t = adminv1.BillingErrorType_BILLING_ERROR_TYPE_NO_PAYMENT_METHOD
			case "trial-ended":
				t = adminv1.BillingErrorType_BILLING_ERROR_TYPE_TRIAL_ENDED
			case "payment-failed":
				t = adminv1.BillingErrorType_BILLING_ERROR_TYPE_PAYMENT_FAILED
			default:
				return fmt.Errorf("invalid error type %q", errorType)
			}

			_, err = client.SudoDeleteOrganizationBillingError(ctx, &adminv1.SudoDeleteOrganizationBillingErrorRequest{
				Organization: org,
				Type:         t,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Billing error of type %q deleted for organization %q\n", errorType, org)

			return nil
		},
	}

	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&errorType, "type", "", "Billing Error Type")
	return setCmd
}
