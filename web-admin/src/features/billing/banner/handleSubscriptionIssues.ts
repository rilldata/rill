import { V1BillingIssueType } from "@rilldata/web-admin/client";
import type { V1BillingIssue } from "@rilldata/web-admin/client";
import type { ShowTeamPlanDialogCallback } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";
import { DateTime } from "luxon";

export function getNeverSubscribedIssue(issues: V1BillingIssue[]) {
  return issues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_NEVER_SUBSCRIBED,
  );
}
export function getCancelledIssue(issues: V1BillingIssue[]) {
  return issues.find(
    (i) =>
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED,
  );
}

export function handleSubscriptionIssues(
  cancelledSubIssue: V1BillingIssue,
  onShowStartTeamPlan: ShowTeamPlanDialogCallback,
) {
  let accessTimeout = "";

  if (cancelledSubIssue.metadata.subscriptionCancelled?.endDate) {
    const endDate = DateTime.fromJSDate(
      new Date(cancelledSubIssue.metadata.subscriptionCancelled?.endDate),
    );
    if (endDate.isValid && endDate.millisecond > Date.now()) {
      accessTimeout = ` but you still have access through ${endDate.toLocaleString(DateTime.DATE_MED)}`;
    }
  }

  return <BannerMessage>{
    type: "warning",
    message: `Your plan was canceled${accessTimeout}. To maintain access, renew your plan.`,
    iconType: "alert",
    cta: {
      text: "Renew ->",
      type: "button",
      onClick: () => {
        onShowStartTeamPlan(
          "renew",
          cancelledSubIssue.metadata.subscriptionCancelled?.endDate ?? "",
        );
      },
    },
  };
}

export function cancelledSubscriptionHasEnded(
  cancelledSubIssue: V1BillingIssue,
) {
  const endDate = new Date(
    cancelledSubIssue.metadata?.subscriptionCancelled?.endDate,
  );
  const endTime = endDate.getTime();
  return Number.isNaN(endTime) || endTime < Date.now();
}
