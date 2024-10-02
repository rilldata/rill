import {
  createAdminServiceListOrganizationBillingIssues,
  type V1BillingIssue,
  V1BillingIssueType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";

export const PaymentBillingIssueTypes: Partial<
  Record<V1BillingIssueType, string>
> = {
  [V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED]:
    "Payment method has failed.",
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD]:
    "There is no payment method on file.",
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS]:
    "There is no billing address on file.",
};

export function getPaymentIssues(organization: string) {
  return createAdminServiceListOrganizationBillingIssues(organization, {
    query: {
      select: (data) =>
        data.issues?.filter((i) => i.type in PaymentBillingIssueTypes),
    },
  });
}

export function handlePaymentIssues(
  subscription: V1Subscription,
  issues: V1BillingIssue[],
) {
  const issue = issues[0];
  const bannerMessage: BannerMessage = {
    type: "warning",
    message: PaymentBillingIssueTypes[issue.type],
    iconType: "alert",
  };

  return bannerMessage;
}
