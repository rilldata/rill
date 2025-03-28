package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const (
	PlanStarter    = "starter"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// AllPlans is a list of all available plans
var AllPlans = []string{PlanStarter, PlanPro, PlanEnterprise}

func SubscriptionCmd(ch *cmdutil.Helper) *cobra.Command {
	subsCmd := &cobra.Command{
		Use:               "subscription",
		Short:             "Manage organisation subscription",
		PersistentPreRunE: cmdutil.CheckAuth(ch),
	}

	subsCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	subsCmd.AddCommand(EditCmd(ch))
	subsCmd.AddCommand(ListCmd(ch))
	subsCmd.AddCommand(CancelCmd(ch))
	subsCmd.AddCommand(RenewCmd(ch))

	return subsCmd
}

// IsValidPlan checks if the given plan is valid
func IsValidPlan(plan string) bool {
	for _, p := range AllPlans {
		if p == plan {
			return true
		}
	}
	return false
}
