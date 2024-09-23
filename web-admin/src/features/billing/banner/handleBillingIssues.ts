import type {
  V1BillingIssue,
  V1BillingPlan,
  V1Subscription,
} from "@rilldata/web-admin/client";
import {
  handlePaymentIssues,
  PaymentBillingIssueTypes,
} from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
import { handleTrialPlan } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export function handleBillingIssues(
  subscription: V1Subscription,
  plan: V1BillingPlan,
  issues: V1BillingIssue[],
) {
  if (plan.trialPeriodDays) {
    eventBus.emit("banner", handleTrialPlan(subscription, issues));
    return;
  }

  const paymentIssues = issues.filter(
    (i) => i.type in PaymentBillingIssueTypes,
  );
  if (paymentIssues.length) {
    eventBus.emit("banner", handlePaymentIssues(subscription, paymentIssues));
    return;
  }
}
