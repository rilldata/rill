import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";

export function getCreditIssue(issues: V1BillingIssue[]) {
  return {
    creditLow: issues.find(
      (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_LOW,
    ),
    creditCritical: issues.find(
      (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_CRITICAL,
    ),
    creditExhausted: issues.find(
      (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_EXHAUSTED,
    ),
  };
}

export function getMessageForCreditIssue(
  issue: V1BillingIssue,
): BillingIssueMessage {
  switch (issue.type) {
    case V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_LOW: {
      const remaining = issue.metadata?.creditLow?.creditRemaining ?? 0;
      const total = issue.metadata?.creditLow?.creditTotal ?? 250;
      const pct = Math.round(((total - remaining) / total) * 100);
      return {
        type: "warning",
        iconType: "alert",
        title: `You've used ${pct}% of your $${total} free credit.`,
        description: `$${remaining.toFixed(0)} remaining — upgrade to Growth to keep your projects running.`,
        cta: {
          type: "upgrade",
          text: "Upgrade to Growth",
          growthPlanDialogType: "credit-low",
        },
      };
    }

    case V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_CRITICAL: {
      const remaining = issue.metadata?.creditCritical?.creditRemaining ?? 0;
      return {
        type: "warning",
        iconType: "alert",
        title: "Your free credit is almost exhausted.",
        description: `$${remaining.toFixed(0)} remaining — upgrade now to avoid project hibernation.`,
        cta: {
          type: "upgrade",
          text: "Upgrade to Growth",
          growthPlanDialogType: "credit-low",
        },
      };
    }

    case V1BillingIssueType.BILLING_ISSUE_TYPE_CREDIT_EXHAUSTED:
      return {
        type: "error",
        iconType: "alert",
        title:
          "Your free credit is exhausted and your projects have been hibernated.",
        description:
          "Upgrade to Growth to wake your projects and resume access.",
        cta: {
          type: "upgrade",
          text: "Upgrade to Growth",
          growthPlanDialogType: "credit-exhausted",
        },
      };

    default:
      return {
        type: "default",
        iconType: "alert",
        title: "",
        description: "",
      };
  }
}
