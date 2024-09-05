package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteWarningCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, warningType string
	warnings := []string{"on-trial"}
	setCmd := &cobra.Command{
		Use:   "delete-warning",
		Short: "Delete billing warning of a type for an organization",
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

			if warningType == "" {
				warningType, err = cmdutil.SelectPrompt("Select warning type to delete", warnings, "")
				if err != nil {
					return err
				}
			}

			var w adminv1.BillingWarningType
			switch warningType {
			case "on-trial":
				w = adminv1.BillingWarningType_BILLING_WARNING_TYPE_ON_TRIAL
			default:
				return fmt.Errorf("invalid warning type %q", warningType)
			}

			_, err = client.SudoDeleteOrganizationBillingWarning(ctx, &adminv1.SudoDeleteOrganizationBillingWarningRequest{
				Organization: org,
				Type:         w,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Billing warning of type %q deleted for organization %q\n", warningType, org)

			return nil
		},
	}

	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&warningType, "type", "", "Billing Warning Type")
	return setCmd
}
