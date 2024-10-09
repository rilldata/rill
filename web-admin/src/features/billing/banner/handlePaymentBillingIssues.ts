import {
  type V1BillingIssue,
  V1BillingIssueType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";

export const PaymentBillingIssueTypes: Partial<
  Record<V1BillingIssueType, { long: string; short: string }>
> = {
  [V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED]: {
    long: "Input a valid payment to maintain access.",
    short: "payment",
  },
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD]: {
    long: "Input a valid payment to maintain access.",
    short: "payment",
  },
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS]: {
    long: "Input a valid billing address to maintain access.",
    short: "billing address",
  },
};

export function getPaymentIssues(issues: V1BillingIssue[]) {
  return issues?.filter((i) => i.type in PaymentBillingIssueTypes);
}

export function getPaymentIssueErrorText(paymentIssues: V1BillingIssue[]) {
  const paymentFailed = paymentIssues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED,
  );
  if (paymentFailed) {
    return "Payment has failed.";
  }

  const issueTexts = paymentIssues.map(
    (i) => PaymentBillingIssueTypes[i.type].short,
  );
  return `No valid ${issueTexts.join(" or ")}.`;
}

export function handlePaymentIssues(
  organization: string,
  subscription: V1Subscription,
  issues: V1BillingIssue[],
) {
  const issue = issues[0];
  const bannerMessage: BannerMessage = {
    type: "warning",
    message: PaymentBillingIssueTypes[issue.type].long,
    iconType: "alert",
    cta: {
      text: "Update payment methods ->",
      type: "button",
      onClick: async () => {
        window.open(
          await fetchPaymentsPortalURL(organization, window.location.href),
          "_self",
        );
      },
    },
  };

  return bannerMessage;
}
