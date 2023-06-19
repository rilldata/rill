package quota

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(cfg *config.Config) *cobra.Command {
	var org, email string
	getCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.NoArgs,
		Short: "Get quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if org != "" {
				res, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{
					Name: org,
				})
				if err != nil {
					return err
				}

				orgQuotas := res.Organization

				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("QuotaProjects: %d\n", orgQuotas.Quotas.QuotaProjects)
				fmt.Printf("QuotaDeployments: %d\n", orgQuotas.Quotas.QuotaDeployments)
				fmt.Printf("QuotaSlotsTotal: %d\n", orgQuotas.Quotas.QuotaSlotsTotal)
				fmt.Printf("QuotaSlotsPerDeployment: %d\n", orgQuotas.Quotas.QuotaSlotsPerDeployment)
				fmt.Printf("QuotaOutstandingInvites: %d\n", orgQuotas.Quotas.QuotaOutstandingInvites)
			} else if email != "" {
				res, err := client.GetUser(ctx, &adminv1.GetUserRequest{
					Email: email,
				})
				if err != nil {
					return err
				}

				userQuotas := res.User
				fmt.Printf("User: %s\n", email)
				fmt.Printf("QuotaProjects: %d\n", userQuotas.Quotas.QuotaSingleuserOrgs)
			} else {
				return fmt.Errorf("Pleasr provide org|user")
			}

			return nil
		},
	}

	getCmd.Flags().SortFlags = false
	getCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	getCmd.Flags().StringVar(&email, "user", "", "User Email")
	return getCmd
}
