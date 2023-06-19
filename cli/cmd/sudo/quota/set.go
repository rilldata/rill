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
				req := &adminv1.SudoUpdateOrganizationQuotasRequest{
					OrgName: org,
				}

				switch key {
				case "quota_projects":
					req.Quota = &adminv1.SudoUpdateOrganizationQuotasRequest_QuotaProjects{
						QuotaProjects: uint32(value),
					}
				case "quota_deployments":
					req.Quota = &adminv1.SudoUpdateOrganizationQuotasRequest_QuotaDeployments{
						QuotaDeployments: uint32(value),
					}
				case "quota_slots_total":
					req.Quota = &adminv1.SudoUpdateOrganizationQuotasRequest_QuotaSlotsTotal{
						QuotaSlotsTotal: uint32(value),
					}
				case "quota_slots_per_deployment":
					req.Quota = &adminv1.SudoUpdateOrganizationQuotasRequest_QuotaSlotsPerDeployment{
						QuotaSlotsPerDeployment: uint32(value),
					}
				case "quota_outstanding_invites":
					req.Quota = &adminv1.SudoUpdateOrganizationQuotasRequest_QuotaOutstandingInvites{
						QuotaOutstandingInvites: uint32(value),
					}
				default:
					return fmt.Errorf("invalid quota key %q", args[0])
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
				res, err := client.SudoUpdateUserQuotas(ctx, &adminv1.SudoUpdateUserQuotasRequest{
					Email:               email,
					QuotaSingleuserOrgs: uint32(value),
				})
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
	return setCmd
}
