import {
  adminServiceListOrganizationBillingIssues,
  createAdminServiceListOrganizationBillingIssues,
  getAdminServiceListOrganizationBillingIssuesQueryKey,
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import {
  cancelledSubscriptionHasEnded,
  getCancelledIssue,
  getNeverSubscribedIssue,
} from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import {
  getTrialIssue,
  trialHasPastGracePeriod,
} from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
import { getCreditIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCreditIssues";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function fetchOrganizationBillingIssues(organization: string) {
  const resp = await queryClient.fetchQuery({
    queryKey:
      getAdminServiceListOrganizationBillingIssuesQueryKey(organization),
    queryFn: () => adminServiceListOrganizationBillingIssues(organization),
    staleTime: Infinity,
  });
  return resp.issues ?? [];
}

export type CategorisedOrganizationBillingIssues = {
  neverSubscribed?: V1BillingIssue;
  trial?: V1BillingIssue;
  cancelled?: V1BillingIssue;
  payment: V1BillingIssue[];
  creditLow?: V1BillingIssue;
  creditCritical?: V1BillingIssue;
  creditExhausted?: V1BillingIssue;
};
export function useCategorisedOrganizationBillingIssues(organization: string) {
  return createAdminServiceListOrganizationBillingIssues(
    organization,
    undefined,
    {
      query: {
        select: (data) => {
          const issues = data.issues ?? [];
          const credit = getCreditIssue(issues);
          return <CategorisedOrganizationBillingIssues>{
            neverSubscribed: getNeverSubscribedIssue(issues),
            trial: getTrialIssue(issues),
            cancelled: getCancelledIssue(issues),
            payment: getPaymentIssues(issues),
            creditLow: credit.creditLow,
            creditCritical: credit.creditCritical,
            creditExhausted: credit.creditExhausted,
          };
        },
      },
    },
  );
}

export function hasBlockerIssues(issues: V1BillingIssue[]) {
  // Credit exhaustion is a blocker
  const credit = getCreditIssue(issues);
  if (credit.creditExhausted) return true;

  const trialIssue = getTrialIssue(issues);
  if (trialIssue) {
    return (
      trialIssue.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED &&
      trialHasPastGracePeriod(trialIssue)
    );
  }

  const subCancelled = getCancelledIssue(issues);
  if (subCancelled) return cancelledSubscriptionHasEnded(subCancelled);

  const payment = getPaymentIssues(issues);
  return !!payment.length;
}
