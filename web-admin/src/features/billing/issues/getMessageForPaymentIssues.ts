import {
  type V1BillingIssue,
  type V1BillingIssueMetadataPaymentFailedMeta,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";

export const PaymentBillingIssueTypes: Partial<
  Record<V1BillingIssueType, { long: string; short: string }>
> = {
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD]: {
    long: "Input a valid payment to maintain access.",
    short: "payment method",
  },
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS]: {
    long: "Input a valid billing address to maintain access.",
    short: "billing address",
  },
};

export function getPaymentIssues(issues: V1BillingIssue[]) {
  return issues?.filter(
    (i) =>
      i.type &&
      (i.type in PaymentBillingIssueTypes ||
        i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED),
  );
}

export function getPaymentIssueErrorText(paymentIssues: V1BillingIssue[]) {
  const issueTexts = paymentIssues
    .map((i) => PaymentBillingIssueTypes[i.type ?? ""]?.short)
    .filter(Boolean) as string[];

  return `No valid ${issueTexts.length ? issueTexts.join(" or ") : "payment method"} on file.`;
}

export function getMessageForPaymentIssues(
  isCustomPlan: boolean,
  issues: V1BillingIssue[],
) {
  const paymentFailed = issues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_PAYMENT_FAILED,
  );
  if (!paymentFailed) {
    // no need to handle anything if payment has not yet failed
    return undefined;
  }

  const oldestInvoice = findOldestInvoice(paymentFailed);
  if (!oldestInvoice || !invoiceIsDue(oldestInvoice)) {
    return undefined;
  }

  const overdue = invoiceIsOverdue(oldestInvoice);

  const issueTexts = issues
    .map((i) => PaymentBillingIssueTypes[i.type ?? ""]?.short)
    .filter(Boolean) as string[];

  const message: BillingIssueMessage = {
    type: overdue ? "error" : "warning",
    title: "",
    description: "",
    iconType: "alert",
  };
  const overdueTitleSuffix = overdue
    ? " and this org’s projects are now hibernating"
    : "";
  if (isCustomPlan) {
    message.title = `Your invoice is past due${overdueTitleSuffix}.`;
    message.description = overdue
      ? "Contact us to regain access."
      : "To maintain access, please contact us.";
    message.cta = {
      type: "contact",
      text: "Contact us",
    };
  } else {
    message.title = `Your subscription is past due${overdueTitleSuffix}.`;
    message.description =
      `Input a valid ${issueTexts.length ? issueTexts.join(" or ") : "payment method"} ` +
      (overdue
        ? "to wake projects and regain full access."
        : "to maintain access.");
    message.cta = {
      type: "payment",
      text: "Update payment methods",
    };
  }

  return message;
}

function findOldestInvoice(paymentFailed: V1BillingIssue) {
  if (!paymentFailed.metadata?.paymentFailed?.invoices?.length)
    return undefined;

  let oldest = paymentFailed.metadata.paymentFailed.invoices[0];
  for (
    let i = 1;
    i < paymentFailed.metadata.paymentFailed.invoices.length;
    i++
  ) {
    const invoice = paymentFailed.metadata.paymentFailed.invoices[i];
    if (
      new Date(invoice.dueDate ?? "").getTime() >
      new Date(oldest.dueDate ?? "").getTime()
    ) {
      oldest = invoice;
    }
  }

  return oldest;
}

function invoiceIsDue(invoice: V1BillingIssueMetadataPaymentFailedMeta) {
  const dueDate = new Date(invoice.dueDate ?? "");
  const dueDateTime = dueDate.getTime();
  return Number.isNaN(dueDateTime) || dueDateTime < Date.now();
}

function invoiceIsOverdue(invoice: V1BillingIssueMetadataPaymentFailedMeta) {
  const gracePeriod = new Date(invoice.gracePeriodEndDate ?? "");
  const gracePeriodTime = gracePeriod.getTime();
  return Number.isNaN(gracePeriodTime) || gracePeriodTime < Date.now();
}
