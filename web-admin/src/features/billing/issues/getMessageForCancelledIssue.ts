import { V1BillingIssueType } from "@rilldata/web-admin/client";
import type { V1BillingIssue } from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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

export function getMessageForCancelledIssue(cancelledSubIssue: V1BillingIssue) {
  let accessTimeout = "";
  let ended = false;

  if (cancelledSubIssue.metadata?.subscriptionCancelled?.endDate) {
    const endDate = DateTime.fromJSDate(
      new Date(cancelledSubIssue.metadata.subscriptionCancelled.endDate),
    );
    if (endDate.isValid && endDate.toMillis() > Date.now()) {
      accessTimeout = m.billing_access_until({ date: endDate.toLocaleString(DateTime.DATE_MED) });
    }
  }
  if (!accessTimeout) {
    accessTimeout = m.billing_subscription_ended();
    ended = true;
  }

  return <BillingIssueMessage>{
    type: ended ? "error" : "warning",
    title: m.billing_plan_cancelled({ accessTimeout }),
    description: m.billing_renew_to_maintain_access(),
    iconType: "alert",
    cta: {
      text: m.billing_renew(),
      type: "show-upgrade",
      teamPlanDialogType: "renew",
      teamPlanEndDate:
        cancelledSubIssue.metadata?.subscriptionCancelled?.endDate,
    },
  };
}

export function cancelledSubscriptionHasEnded(
  cancelledSubIssue: V1BillingIssue,
) {
  const endDate = new Date(
    cancelledSubIssue.metadata?.subscriptionCancelled?.endDate ?? "",
  );
  const endTime = endDate.getTime();
  return Number.isNaN(endTime) || endTime < Date.now();
}
