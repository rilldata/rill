import {
  type V1BillingIssue,
  type V1BillingIssueMetadataPaymentFailedMeta,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export const PaymentBillingIssueTypes: Partial<
  Record<V1BillingIssueType, true>
> = {
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD]: true,
  [V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS]: true,
};

function getPaymentIssueShortText(type: V1BillingIssueType | undefined): string {
  switch (type) {
    case V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD:
      return m.billing_payment_method_short();
    case V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS:
      return m.billing_billing_address_short();
    default:
      return "";
  }
}

export function needsPaymentSetup(issues: V1BillingIssue[]): boolean {
  const hasNoPaymentMethodIssue = issues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD,
  );
  const hasNoBillingAddressIssue = issues.find(
    (i) => i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS,
  );
  return Boolean(hasNoPaymentMethodIssue && hasNoBillingAddressIssue);
}

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
    .map((i) => getPaymentIssueShortText(i.type))
    .filter(Boolean) as string[];

  const methods = issueTexts.length
    ? issueTexts.join(` ${m.billing_or()} `)
    : m.billing_payment_method_short();
  return m.billing_no_valid_on_file({ methods });
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
    .map((i) => getPaymentIssueShortText(i.type))
    .filter(Boolean) as string[];

  const message: BillingIssueMessage = {
    type: overdue ? "error" : "warning",
    title: "",
    description: "",
    iconType: "alert",
  };
  const overdueTitleSuffix = overdue
    ? m.billing_projects_hibernating()
    : "";
  if (isCustomPlan) {
    message.title = m.billing_invoice_past_due({ suffix: overdueTitleSuffix });
    message.description = overdue
      ? m.billing_contact_us_to_regain()
      : m.billing_contact_us_to_maintain();
    message.cta = {
      type: "contact",
      text: m.billing_contact_us_cta(),
    };
  } else {
    message.title = m.billing_subscription_past_due({ suffix: overdueTitleSuffix });
    const methods = issueTexts.length
      ? issueTexts.join(` ${m.billing_or()} `)
      : m.billing_payment_method_short();
    message.description = overdue
      ? m.billing_input_valid_to_wake({ methods })
      : m.billing_input_valid_to_maintain({ methods });
    message.cta = {
      type: "payment",
      text: m.billing_update_payment_methods(),
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
