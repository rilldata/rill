import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
import { handlePaymentIssues } from "@rilldata/web-admin/features/billing/issues/handlePaymentBillingIssues";
import { handleSubscriptionIssues } from "@rilldata/web-admin/features/billing/issues/handleSubscriptionIssues";
import { handleTrialPlan } from "@rilldata/web-admin/features/billing/issues/handleTrialPlan";
import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
import {
  getSubscriptionForOrg,
  useCategorisedOrganizationBillingIssues,
} from "@rilldata/web-admin/features/billing/selectors";
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
  type: "upgrade" | "payment" | "wake-projects";
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
      getSubscriptionForOrg(organization),
      useCategorisedOrganizationBillingIssues(organization),
    ],
    ([orgResp, subscriptionResp, categorisedIssuesResp]) => {
      if (
        orgResp.isFetching ||
        (!orgResp.data?.permissions?.manageOrg &&
          subscriptionResp.isFetching) ||
        categorisedIssuesResp.isFetching
      ) {
        return {
          isFetching: true,
        };
      }
      if (
        orgResp.error ||
        subscriptionResp.error ||
        categorisedIssuesResp.error
      ) {
        return {
          isFetching: false,
          error:
            orgResp.error ??
            subscriptionResp.error ??
            categorisedIssuesResp.error,
        };
      }

      if (categorisedIssuesResp.data.cancelled) {
        return {
          isFetching: false,
          data: handleSubscriptionIssues(categorisedIssuesResp.data.cancelled),
        };
      }

      if (categorisedIssuesResp.data.trial) {
        return {
          isFetching: false,
          data: handleTrialPlan(categorisedIssuesResp.data.trial),
        };
      }

      if (
        categorisedIssuesResp.data.payment.length &&
        subscriptionResp.data?.subscription
      ) {
        return {
          isFetching: false,
          data: handlePaymentIssues(
            organization,
            categorisedIssuesResp.data.payment,
          ),
        };
      }

      return {
        isFetching: false,
      };
    },
  );
}
