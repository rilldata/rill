import type {
  V1BillingIssue,
  V1Subscription,
} from "@rilldata/web-admin/client";
import {
  handlePaymentIssues,
  PaymentBillingIssueTypes,
} from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
import { handleTrialPlan } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";
import { isTrialPlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export function handleBillingIssues(
  subscription: V1Subscription,
  issues: V1BillingIssue[],
) {
  if (isTrialPlan(subscription.plan)) {
    eventBus.emit("banner", handleTrialPlan(issues));
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
