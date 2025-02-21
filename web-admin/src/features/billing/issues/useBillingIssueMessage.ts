import {
  createAdminServiceGetBillingSubscription,
  createAdminServiceGetOrganization,
} from "@rilldata/web-admin/client";
import { getMessageForPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { getMessageForCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { getMessageForTrialPlan } from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
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
      createAdminServiceGetOrganization(organization),
      useCategorisedOrganizationBillingIssues(organization),
      areAllProjectsHibernating(organization),
    ],
    ([orgResp, categorisedIssuesResp, allProjectsHibernatingResp]) => {
      if (
        orgResp.isFetching ||
        categorisedIssuesResp.isFetching ||
        allProjectsHibernatingResp.isFetching
      ) {
        return {
          isFetching: true,
          error: undefined,
        };
      }
      if (
        orgResp.error ||
        categorisedIssuesResp.error ||
        allProjectsHibernatingResp.error
      ) {
        return {
          isFetching: false,
          error:
            orgResp.error ??
            categorisedIssuesResp.error ??
            allProjectsHibernatingResp.error,
        };
      }

      if (categorisedIssuesResp.data?.cancelled) {
        return {
          isFetching: false,
          error: undefined,
          data: getMessageForCancelledIssue(
            categorisedIssuesResp.data.cancelled,
          ),
        };
      }

      if (categorisedIssuesResp.data?.trial) {
        return {
          isFetching: false,
          error: undefined,
          data: getMessageForTrialPlan(categorisedIssuesResp.data.trial),
        };
      }

      if (
        categorisedIssuesResp.data?.payment.length &&
        orgResp.data?.organization?.billingPlanType
      ) {
        const paymentIssue = getMessageForPaymentIssues(
          !isTeamPlan(orgResp.data.organization.billingPlanType),
          categorisedIssuesResp.data.payment,
        );
        // if we do not have any payment related message to show, skip it
        if (paymentIssue)
          return {
            isFetching: false,
            error: undefined,
            data: paymentIssue,
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
