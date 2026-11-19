import {
  createAdminServiceGetBillingCreditBalance,
  createAdminServiceGetOrganization,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import { getMessageForPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { getMessageForCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { getMessageForCustomMessage } from "@rilldata/web-admin/features/billing/issues/getMessageForCustomMessage";
import { getMessageForTrialPlan } from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
import {
  isProPlan,
  isTeamPlan,
} from "@rilldata/web-admin/features/billing/plans/utils";
import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived } from "svelte/store";

export type BillingIssueMessage = {
  type: BannerMessage["type"];
  iconType: BannerMessage["iconType"];
  title: string;
  description: string;
  // Whether title or description contain HTML, e.g. links.
  includesHtml?: boolean;
  cta?: BillingIssueMessageCTA;
  dismissible?: BannerMessage["dismissible"];
};
export type BillingIssueMessageCTA = {
  type: "show-upgrade" | "payment" | "contact";
  text: string;

  teamPlanDialogType?: TeamPlanDialogTypes;
  teamPlanEndDate?: string;
};

export function useBillingIssueMessage(organization: string) {
  const categorisedIssues =
    useCategorisedOrganizationBillingIssues(organization);
  return derived(
    [
      createAdminServiceGetOrganization(organization),
      categorisedIssues,
      useTrialCreditBalance(organization, categorisedIssues),
    ],
    ([orgResp, categorisedIssuesResp, trialCreditBalance]) => {
      if (
        orgResp.isLoading ||
        categorisedIssuesResp.isLoading ||
        trialCreditBalance.isLoading
      ) {
        return {
          isFetching: true,
          isLoading: true,
          error: undefined,
        };
      }
      if (orgResp.error || categorisedIssuesResp.error) {
        return {
          isFetching: false,
          isLoading: false,
          error: orgResp.error ?? categorisedIssuesResp.error,
        };
      }

      // A support-set custom message takes precedence over billing-derived messages.
      if (categorisedIssuesResp.data?.message) {
        return {
          isFetching: false,
          isLoading: false,
          error: undefined,
          data: getMessageForCustomMessage(categorisedIssuesResp.data.message),
        };
      }

      if (categorisedIssuesResp.data?.trial) {
        return {
          isFetching: false,
          isLoading: false,
          error: undefined,
          data: getMessageForTrialPlan(
            categorisedIssuesResp.data.trial,
            trialCreditBalance.balance,
          ),
        };
      }

      if (categorisedIssuesResp.data?.cancelled) {
        return {
          isFetching: false,
          isLoading: false,
          error: undefined,
          data: getMessageForCancelledIssue(
            categorisedIssuesResp.data.cancelled,
          ),
        };
      }

      if (
        categorisedIssuesResp.data?.payment.length &&
        orgResp.data?.organization?.billingPlanName
      ) {
        const paymentIssue = getMessageForPaymentIssues(
          !isTeamPlan(orgResp.data.organization.billingPlanName) &&
            !isProPlan(orgResp.data.organization.billingPlanName),
          categorisedIssuesResp.data.payment,
        );
        // if we do not have any payment related message to show, skip it
        if (paymentIssue)
          return {
            isFetching: false,
            isLoading: false,
            error: undefined,
            data: paymentIssue,
          };
      }

      return {
        isFetching: false,
        isLoading: false,
        error: undefined,
      };
    },
  );
}

/**
 * Fetches the remaining credit balance for orgs on the credit trial, for display in the trial banner.
 * Allows a recently cached balance so the banner doesn't query the billing provider on every page load.
 * The balance is undefined for other orgs, or if it could not be fetched.
 */
function useTrialCreditBalance(
  organization: string,
  categorisedIssues: ReturnType<typeof useCategorisedOrganizationBillingIssues>,
) {
  return derived(
    categorisedIssues,
    (categorisedIssuesResp, set) => {
      if (
        categorisedIssuesResp.data?.trial?.type !==
        V1BillingIssueType.BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL
      ) {
        set({ isLoading: false, balance: undefined });
        return;
      }
      return createAdminServiceGetBillingCreditBalance(
        organization,
        { allowCached: true },
        // The server already falls back to a cached balance, so retrying would only delay the banner.
        { query: { retry: false } },
        queryClient,
      ).subscribe((balanceResp) =>
        set({
          isLoading: balanceResp.isLoading,
          // A zero balance is omitted from the response.
          balance: balanceResp.data
            ? (balanceResp.data.balance ?? 0)
            : undefined,
        }),
      );
    },
    { isLoading: true, balance: undefined as number | undefined },
  );
}
