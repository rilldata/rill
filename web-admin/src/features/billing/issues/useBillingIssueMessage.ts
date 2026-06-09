import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
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
import { derived } from "svelte/store";

export type BillingIssueMessage = {
  type: BannerMessage["type"];
  iconType: BannerMessage["iconType"];
  title: string;
  description: string;
  cta?: BillingIssueMessageCTA;
  dismissible?: BannerMessage["dismissible"];
};
export type BillingIssueMessageCTA = {
  type: "upgrade" | "show-upgrade" | "payment" | "contact";
  text: string;

  teamPlanDialogType?: TeamPlanDialogTypes;
  teamPlanEndDate?: string;
};

export function useBillingIssueMessage(organization: string) {
  return derived(
    [
      createAdminServiceGetOrganization(organization),
      useCategorisedOrganizationBillingIssues(organization),
    ],
    ([orgResp, categorisedIssuesResp]) => {
      if (orgResp.isLoading || categorisedIssuesResp.isLoading) {
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
          data: getMessageForTrialPlan(categorisedIssuesResp.data.trial),
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
