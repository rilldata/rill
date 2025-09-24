package quota

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, email string
	var singleUser, trialOrgs, projects, deployments, slotsTotal, slotsPerDeployment, outstandingInvites, numUsers int32
	var storageLimitBytesPerDeployment int64
	setCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.NoArgs,
		Short: "Set quota for user or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if org != "" {
				req := &adminv1.SudoUpdateOrganizationQuotasRequest{
					Org: org,
				}

				if cmd.Flags().Changed("projects") {
					req.Projects = &projects
				}
				if cmd.Flags().Changed("deployments") {
					req.Deployments = &deployments
				}
				if cmd.Flags().Changed("slots-total") {
					req.SlotsTotal = &slotsTotal
				}
				if cmd.Flags().Changed("slots-per-deployment") {
					req.SlotsPerDeployment = &slotsPerDeployment
				}
				if cmd.Flags().Changed("outstanding-invites") {
					req.OutstandingInvites = &outstandingInvites
				}
				if cmd.Flags().Changed("storage-limit-bytes-per-deployment") {
					req.StorageLimitBytesPerDeployment = &storageLimitBytesPerDeployment
				}

				res, err := client.SudoUpdateOrganizationQuotas(ctx, req)
				if err != nil {
					return err
				}

				orgQuotas := res.Organization.Quotas
				ch.PrintfSuccess("Updated organizations quota\n")
				fmt.Printf("Organization Name: %s\n", org)
				fmt.Printf("Projects: %d\n", orgQuotas.Projects)
				fmt.Printf("Deployments: %d\n", orgQuotas.Deployments)
				fmt.Printf("Slots total: %d\n", orgQuotas.SlotsTotal)
				fmt.Printf("Slots per deployment: %d\n", orgQuotas.SlotsPerDeployment)
				fmt.Printf("Outstanding invites: %d\n", orgQuotas.OutstandingInvites)
				fmt.Printf("Storage limit bytes per deployment: %d\n", orgQuotas.StorageLimitBytesPerDeployment)
			} else if email != "" {
				req := &adminv1.SudoUpdateUserQuotasRequest{
					Email: email,
				}

				if cmd.Flags().Changed("singleuser-orgs") {
					req.SingleuserOrgs = &singleUser
				}

				if cmd.Flags().Changed("trial-orgs") {
					req.TrialOrgs = &trialOrgs
				}

				res, err := client.SudoUpdateUserQuotas(ctx, req)
				if err != nil {
					return err
				}

				userQuotas := res.User.Quotas
				ch.PrintfSuccess("Updated user's quota\n")
				fmt.Printf("User: %s\n", email)
				fmt.Printf("Trial orgs: %d\n", userQuotas.TrialOrgs)
				fmt.Printf("Single user orgs: %d\n", userQuotas.SingleuserOrgs)
			} else {
				return fmt.Errorf("Please set --org or --user")
			}

			return nil
		},
	}

	setCmd.Flags().SortFlags = false
	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&email, "user", "", "User Email")
	setCmd.Flags().Int32Var(&singleUser, "singleuser-orgs", 0, "Quota single user org")
	setCmd.Flags().Int32Var(&trialOrgs, "trial-orgs", 0, "Quota trial orgs for a user")
	setCmd.Flags().Int32Var(&projects, "projects", 0, "Quota projects")
	setCmd.Flags().Int32Var(&deployments, "deployments", 0, "Quota deployments")
	setCmd.Flags().Int32Var(&slotsTotal, "slots-total", 0, "Quota slots total")
	setCmd.Flags().Int32Var(&slotsPerDeployment, "slots-per-deployment", 0, "Quota slots per deployment")
	setCmd.Flags().Int32Var(&outstandingInvites, "outstanding-invites", 0, "Quota outstanding invites")
	setCmd.Flags().Int32Var(&numUsers, "num-users", 0, "Number of users")
	setCmd.Flags().Int64Var(&storageLimitBytesPerDeployment, "storage-limit-bytes-per-deployment", 0, "Quota storage limit bytes per deployment")
	return setCmd
}
