import {
  type V1BillingIssue,
  V1BillingIssueLevel,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";

export function getCustomMessageIssue(issues: V1BillingIssue[]) {
  return issues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_MESSAGE,
  );
}

export function getMessageForCustomMessage(
  messageIssue: V1BillingIssue,
): BillingIssueMessage {
  return {
    type:
      messageIssue.level === V1BillingIssueLevel.BILLING_ISSUE_LEVEL_ERROR
        ? "error"
        : "warning",
    title: "",
    description: messageIssue.metadata?.message?.message ?? "",
    iconType: "alert",
  };
}
