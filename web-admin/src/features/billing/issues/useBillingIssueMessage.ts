import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
import { getMessageForPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { getMessageForCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
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
};
export type BillingIssueMessageCTA = {
  type: "upgrade" | "payment" | "contact" | "wake-projects";
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
