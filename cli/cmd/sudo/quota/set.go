package quota

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(cfg *config.Config) *cobra.Command {
	var org, email string
	var quotaSingleUser, quotaProjects, quotaDeployments, quotaSlotsTotal, quotaSlotsPerDeployment, quotaOutstandingInvites uint32
	setCmd := &cobra.Command{
		Use:   "set [org|user]",
		Args:  cobra.NoArgs,
		Short: "Set quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if org != "" {
				req := &adminv1.SudoUpdateOrganizationQuotasRequest{
					OrgName: org,
				}

				if cmd.Flags().Changed("quota_projects") {
					req.QuotaProjects = &quotaProjects
				}
				if cmd.Flags().Changed("quota_deployments") {
					req.QuotaDeployments = &quotaDeployments
				}
				if cmd.Flags().Changed("quota_slots_total") {
					req.QuotaSlotsTotal = &quotaSlotsTotal
				}
				if cmd.Flags().Changed("quota_slots_per_deployment") {
					req.QuotaSlotsPerDeployment = &quotaSlotsPerDeployment
				}
				if cmd.Flags().Changed("quota_outstanding_invites") {
					req.QuotaOutstandingInvites = &quotaOutstandingInvites
				}

				res, err := client.SudoUpdateOrganizationQuotas(ctx, req)
				if err != nil {
					return err
				}

				orgQuotas := res.OrganizationQuotas
				cmdutil.PrintlnSuccess("Updated organizations quota")
				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("QuotaProjects: %d\n", orgQuotas.QuotaProjects)
				fmt.Printf("QuotaDeployments: %d\n", orgQuotas.QuotaDeployments)
				fmt.Printf("QuotaSlotsTotal: %d\n", orgQuotas.QuotaSlotsTotal)
				fmt.Printf("QuotaSlotsPerDeployment: %d\n", orgQuotas.QuotaSlotsPerDeployment)
				fmt.Printf("QuotaOutstandingInvites: %d\n", orgQuotas.QuotaOutstandingInvites)
			} else if email != "" {
				req := &adminv1.SudoUpdateUserQuotasRequest{
					Email: email,
				}

				if cmd.Flags().Changed("quota_singleuser_orgs") {
					req.QuotaSingleuserOrgs = &quotaSingleUser
				}

				res, err := client.SudoUpdateUserQuotas(ctx, req)
				if err != nil {
					return err
				}

				userQuotas := res.UserQuotas
				cmdutil.PrintlnSuccess("Updated users quota")
				fmt.Printf("User: %s\n", email)
				fmt.Printf("QuotaProjects: %d\n", userQuotas.QuotaSingleuserOrgs)
			} else {
				return fmt.Errorf("Please set --org or --user")
			}

			return nil
		},
	}

	setCmd.Flags().SortFlags = false
	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&email, "user", "", "User Email")
	setCmd.Flags().Uint32Var(&quotaSingleUser, "quota_singleuser_orgs", 0, "Quota single user org")
	setCmd.Flags().Uint32Var(&quotaProjects, "quota_projects", 0, "Quota projects")
	setCmd.Flags().Uint32Var(&quotaDeployments, "quota_deployments", 0, "Quota deployments")
	setCmd.Flags().Uint32Var(&quotaSlotsTotal, "quota_slots_total", 0, "Quota slots total")
	setCmd.Flags().Uint32Var(&quotaSlotsPerDeployment, "quota_slots_per_deployment", 0, "Quota slots per deployment")
	setCmd.Flags().Uint32Var(&quotaOutstandingInvites, "quota_outstanding_invites", 0, "Quota outstanding invites")
	return setCmd
}
