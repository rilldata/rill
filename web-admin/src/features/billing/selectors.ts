import type { CreateQueryOptions } from "@rilldata/svelte-query";
import {
  createAdminServiceGetBillingSubscription,
  createAdminServiceGetOrganization,
  createAdminServiceListOrganizationBillingIssues,
  type V1BillingIssue,
  V1BillingIssueType,
  type V1GetBillingSubscriptionResponse,
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
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getSubscriptionForOrg<T = V1GetBillingSubscriptionResponse>(
  org: string,
  queryOptions?: CreateQueryOptions<
    V1GetBillingSubscriptionResponse,
    ErrorType<RpcStatus>,
    T // T is the return type of the `select` function
  >,
): CreateQueryResult<T, ErrorType<RpcStatus>> {
  return derived([createAdminServiceGetOrganization(org)], ([orgResp], set) =>
    createAdminServiceGetBillingSubscription(org, {
      query: {
        ...queryOptions,
        enabled:
          (queryOptions && "enabled" in queryOptions
            ? queryOptions.enabled
            : true) &&
          !!orgResp.data?.permissions?.manageOrg &&
          !!org,
        queryClient,
      },
    }).subscribe(set),
  );
}

export function getPlanForOrg(org: string, enabled = true) {
  return derived([createAdminServiceGetOrganization(org)], ([orgResp], set) =>
    createAdminServiceGetBillingSubscription(org, {
      query: {
        enabled: enabled && !!orgResp.data?.permissions?.manageOrg && !!org,
        select: (data) => data.subscription?.plan,
        queryClient,
      },
    }).subscribe(set),
  );
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
      refetchOnWindowFocus: true,
    },
  });
}

export function orgHasBlockerIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(org, {
    query: {
      select: (data) => hasBlockerIssues(data.issues ?? []),
      refetchOnWindowFocus: true,
    },
  });
}
export function hasBlockerIssues(issues: V1BillingIssue[]) {
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
