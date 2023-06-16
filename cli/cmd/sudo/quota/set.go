package quota

import (
	"fmt"
	"strconv"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(cfg *config.Config) *cobra.Command {
	var org, email string
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Args:  cobra.ExactArgs(2),
		Short: "Set quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			key := args[0]
			value, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if org != "" {
				req := &adminv1.SudoSetOrganizationQuotaRequest{
					OrgName: org,
				}

				switch key {
				case "quota_projects":
					req.Quota = &adminv1.SudoSetOrganizationQuotaRequest_QuotaProjects{
						QuotaProjects: uint32(value),
					}
				case "quota_deployments":
					req.Quota = &adminv1.SudoSetOrganizationQuotaRequest_QuotaDeployments{
						QuotaDeployments: uint32(value),
					}
				case "quota_slots_total":
					req.Quota = &adminv1.SudoSetOrganizationQuotaRequest_QuotaSlotsTotal{
						QuotaSlotsTotal: uint32(value),
					}
				case "quota_slots_per_deployment":
					req.Quota = &adminv1.SudoSetOrganizationQuotaRequest_QuotaSlotsPerDeployment{
						QuotaSlotsPerDeployment: uint32(value),
					}
				case "quota_outstanding_invites":
					req.Quota = &adminv1.SudoSetOrganizationQuotaRequest_QuotaOutstandingInvites{
						QuotaOutstandingInvites: uint32(value),
					}
				default:
					return fmt.Errorf("invalid quota key %q", args[0])
				}

				res, err := client.SudoSetOrganizationQuota(ctx, req)
				if err != nil {
					return err
				}

				orgQuota := res.OrganizationQuota
				cmdutil.PrintlnSuccess("Updated organizations quota")
				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("QuotaProjects: %d\n", orgQuota.QuotaProjects)
				fmt.Printf("QuotaDeployments: %d\n", orgQuota.QuotaDeployments)
				fmt.Printf("QuotaSlotsTotal: %d\n", orgQuota.QuotaSlotsTotal)
				fmt.Printf("QuotaSlotsPerDeployment: %d\n", orgQuota.QuotaSlotsPerDeployment)
				fmt.Printf("QuotaOutstandingInvites: %d\n", orgQuota.QuotaOutstandingInvites)
			} else if email != "" {
				res, err := client.SudoSetUserQuota(ctx, &adminv1.SudoSetUserQuotaRequest{
					Email:               email,
					QuotaSingleuserOrgs: uint32(value),
				})
				if err != nil {
					return err
				}

				userQuota := res.UserQuota
				cmdutil.PrintlnSuccess("Updated users quota")
				fmt.Printf("User: %s\n", email)
				fmt.Printf("QuotaProjects: %d\n", userQuota.QuotaSingleuserOrgs)
			} else {
				return fmt.Errorf("Pleasr provide org|user")
			}

			return nil
		},
	}

	setCmd.Flags().SortFlags = false
	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&email, "user", "", "User Email")
	return setCmd
}
