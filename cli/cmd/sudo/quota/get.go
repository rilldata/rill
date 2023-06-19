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
				res, err := client.SudoGetOrganizationQuotas(ctx, &adminv1.SudoGetOrganizationQuotasRequest{
					OrgName: org,
				})
				if err != nil {
					return err
				}

				orgQuotas := res.OrganizationQuotas

				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("QuotaProjects: %d\n", orgQuotas.QuotaProjects)
				fmt.Printf("QuotaDeployments: %d\n", orgQuotas.QuotaDeployments)
				fmt.Printf("QuotaSlotsTotal: %d\n", orgQuotas.QuotaSlotsTotal)
				fmt.Printf("QuotaSlotsPerDeployment: %d\n", orgQuotas.QuotaSlotsPerDeployment)
				fmt.Printf("QuotaOutstandingInvites: %d\n", orgQuotas.QuotaOutstandingInvites)
			} else if email != "" {
				res, err := client.SudoGetUserQuotas(ctx, &adminv1.SudoGetUserQuotasRequest{
					Email: email,
				})
				if err != nil {
					return err
				}

				userQuotas := res.UserQuotas
				fmt.Printf("User: %s\n", email)
				fmt.Printf("QuotaProjects: %d\n", userQuotas.QuotaSingleuserOrgs)
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
