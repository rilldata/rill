import {
  createAdminServiceGetBillingSubscription,
  createAdminServiceListOrganizationBillingIssues,
  V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import {
  getPaymentIssues,
  PaymentBillingIssueTypes,
} from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
import {
  cancelledSubscriptionHasEnded,
  getCancelledIssue,
  getNeverSubscribedIssue,
} from "@rilldata/web-admin/features/billing/banner/handleSubscriptionIssues";
import {
  getTrialIssue,
  trialHasPastGracePeriod,
} from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";

export function getPlanForOrg(org: string, enabled = true) {
  return createAdminServiceGetBillingSubscription(org, {
    query: {
      enabled: enabled && !!org,
      select: (data) => data.subscription?.plan,
    },
  });
}

export type CategorisedOrganizationBillingIssues = {
  neverSubscribed?: V1BillingIssue;
  trial?: V1BillingIssue;
  cancelled?: V1BillingIssue;
  payment: V1BillingIssue[];
};
export function useCategorisedOrganizationBillingIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(org, {
    query: {
      select: (data) => {
        const issues = data.issues ?? [];
        return <CategorisedOrganizationBillingIssues>{
          neverSubscribed: getNeverSubscribedIssue(issues),
          trial: getTrialIssue(issues),
          cancelled: getCancelledIssue(issues),
          payment: getPaymentIssues(issues),
        };
      },
    },
  });
}

export function getOrgBlockerIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(org, {
    query: {
      select: (data) =>
        data.issues
          ?.map((i) => {
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
          })
          .filter(Boolean)?.[0],
    },
  });
}
