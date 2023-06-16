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
				res, err := client.SudoGetOrganizationQuota(ctx, &adminv1.SudoGetOrganizationQuotaRequest{
					OrgName: org,
				})
				if err != nil {
					return err
				}

				orgQuota := res.OrganizationQuota

				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("QuotaProjects: %d\n", orgQuota.QuotaProjects)
				fmt.Printf("QuotaDeployments: %d\n", orgQuota.QuotaDeployments)
				fmt.Printf("QuotaSlotsTotal: %d\n", orgQuota.QuotaSlotsTotal)
				fmt.Printf("QuotaSlotsPerDeployment: %d\n", orgQuota.QuotaSlotsPerDeployment)
				fmt.Printf("QuotaOutstandingInvites: %d\n", orgQuota.QuotaOutstandingInvites)
			} else if email != "" {
				res, err := client.SudoGetUserQuota(ctx, &adminv1.SudoGetUserQuotaRequest{
					Email: email,
				})
				if err != nil {
					return err
				}

				userQuota := res.UserQuota
				fmt.Printf("User: %s\n", email)
				fmt.Printf("QuotaProjects: %d\n", userQuota.QuotaSingleuserOrgs)
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
