import {
  createAdminServiceGetBillingSubscription,
  createAdminServiceListOrganizationBillingIssues,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import { PaymentBillingIssueTypes } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
import { cancelledSubscriptionHasEnded } from "@rilldata/web-admin/features/billing/banner/handleSubscriptionIssues";
import { trialHasPastGracePeriod } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";

export function getPlanForOrg(org: string, enabled = true) {
  return createAdminServiceGetBillingSubscription(org, {
    query: {
      enabled: enabled && !!org,
      select: (data) => data.subscription.plan,
    },
  });
}

export function getOrgBlockerIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(org, {
    query: {
      select: (data) =>
        data.issues?.map((i) => {
          switch (i.type) {
            case V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED:
              return trialHasPastGracePeriod(i) ? "Trial has ended." : "";
            case V1BillingIssueType.BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED:
              return cancelledSubscriptionHasEnded(i)
                ? "Subscription cancelled."
                : "";
            default:
              return i.type in PaymentBillingIssueTypes
                ? "Invoice payment failed"
                : "";
          }
        })?.[0],
    },
  });
}
