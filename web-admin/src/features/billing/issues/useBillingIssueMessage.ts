import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
import { getMessageForPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { getMessageForCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { getMessageForTrialPlan } from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { areAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
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

export function useBillingIssueMessage(
  organization: string,
): CompoundQueryResult<BillingIssueMessage> {
  return derived(
    [
      createAdminServiceGetBillingSubscription(organization),
      useCategorisedOrganizationBillingIssues(organization),
      areAllProjectsHibernating(organization),
    ],
    ([subscriptionResp, categorisedIssuesResp, allProjectsHibernatingResp]) => {
      if (
        subscriptionResp.isFetching ||
        categorisedIssuesResp.isFetching ||
        allProjectsHibernatingResp.isFetching
      ) {
        return {
          isFetching: true,
          error: undefined,
        };
      }
      if (
        subscriptionResp.error ||
        categorisedIssuesResp.error ||
        allProjectsHibernatingResp.error
      ) {
        return {
          isFetching: false,
          error:
            subscriptionResp.error ??
            categorisedIssuesResp.error ??
            allProjectsHibernatingResp.error,
        };
      }

      if (categorisedIssuesResp.data.cancelled) {
        return {
          isFetching: false,
          error: undefined,
          data: getMessageForCancelledIssue(
            categorisedIssuesResp.data.cancelled,
          ),
        };
      }

      if (categorisedIssuesResp.data.trial) {
        return {
          isFetching: false,
          error: undefined,
          data: getMessageForTrialPlan(categorisedIssuesResp.data.trial),
        };
      }

      if (
        categorisedIssuesResp.data.payment.length &&
        subscriptionResp.data?.subscription
      ) {
        return {
          isFetching: false,
          error: undefined,
          data: getMessageForPaymentIssues(
            !!subscriptionResp.data.subscription.plan &&
              !isTeamPlan(subscriptionResp.data.subscription.plan),
            categorisedIssuesResp.data.payment,
          ),
        };
      }

      if (allProjectsHibernatingResp.data) {
        return {
          isFetching: false,
          error: undefined,
          data: <BillingIssueMessage>{
            type: "default",
            title: "Your org’s projects are hibernating.",
            description: "",
            iconType: "sleep",
            cta: {
              type: "wake-projects",
              text: "Wake projects",
            },
          },
        };
      }

      return {
        isFetching: false,
        error: undefined,
      };
    },
  );
}
