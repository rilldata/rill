import {
  createAdminServiceListOrganizationBillingIssues,
  type V1BillingIssue,
  V1BillingIssueType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";

export const PaymentBillingIssueTypes: Partial<
  Record<V1BillingIssueType, string>
> = {
  [V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED]:
    "Input a valid payment to maintain access.",
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD]:
    "Input a valid payment to maintain access.",
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS]:
    "Input a valid billing address to maintain access.",
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
  organization: string,
  subscription: V1Subscription,
  issues: V1BillingIssue[],
) {
  const issue = issues[0];
  const bannerMessage: BannerMessage = {
    type: "warning",
    message: PaymentBillingIssueTypes[issue.type],
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
