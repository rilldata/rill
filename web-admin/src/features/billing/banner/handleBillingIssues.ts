import type { V1Subscription } from "@rilldata/web-admin/client";
import { handlePaymentIssues } from "@rilldata/web-admin/features/billing/banner/handlePaymentBillingIssues";
import { handleSubscriptionIssues } from "@rilldata/web-admin/features/billing/banner/handleSubscriptionIssues";
import { handleTrialPlan } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";
import type { ShowTeamPlanDialogCallback } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
import type { CategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export function handleBillingIssues(
  organization: string,
  subscription: V1Subscription | undefined,
  categorisedIssues: CategorisedOrganizationBillingIssues,
  onShowStartTeamPlan: ShowTeamPlanDialogCallback,
) {
  if (categorisedIssues.cancelled) {
    eventBus.emit(
      "banner",
      handleSubscriptionIssues(
        categorisedIssues.cancelled,
        onShowStartTeamPlan,
      ),
    );
    return;
  }

  if (categorisedIssues.trial) {
    eventBus.emit(
      "banner",
      handleTrialPlan(categorisedIssues.trial, onShowStartTeamPlan),
    );
    return;
  }

  if (categorisedIssues.payment.length && subscription) {
    eventBus.emit(
      "banner",
      handlePaymentIssues(
        organization,
        subscription,
        categorisedIssues.payment,
      ),
    );
    return;
  }

  eventBus.emit("banner", {
    type: "clear",
    iconType: "none",
    message: "",
  });
}
