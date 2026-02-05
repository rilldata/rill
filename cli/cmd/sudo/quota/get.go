package quota

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, email string
	getCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.NoArgs,
		Short: "Get quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if org != "" {
				res, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{
					Org:                  org,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return err
				}

				orgQuotas := res.Organization.Quotas

				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("Projects: %d\n", orgQuotas.Projects)
				fmt.Printf("Deployments: %d\n", orgQuotas.Deployments)
				fmt.Printf("Slots total: %d\n", orgQuotas.SlotsTotal)
				fmt.Printf("Slots per deployment: %d\n", orgQuotas.SlotsPerDeployment)
				fmt.Printf("Outstanding invites: %d\n", orgQuotas.OutstandingInvites)
				fmt.Printf("Storage limit bytes per deployment: %d\n", orgQuotas.StorageLimitBytesPerDeployment)
			} else if email != "" {
				res, err := client.GetUser(ctx, &adminv1.GetUserRequest{
					Email: email,
				})
				if err != nil {
					return err
				}

				userQuotas := res.User.Quotas
				fmt.Printf("User: %s\n", email)
				fmt.Printf("Projects: %d\n", userQuotas.SingleuserOrgs)
				fmt.Printf("Trial Orgs: %d\n", userQuotas.TrialOrgs)
			} else {
				return fmt.Errorf("Please set --org or --user")
			}

			return nil
		},
	}

	getCmd.Flags().SortFlags = false
	getCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	getCmd.Flags().StringVar(&email, "user", "", "User Email")
	return getCmd
}
