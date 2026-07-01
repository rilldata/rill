import { type CategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
import {
  fetchPaymentsPortalURL,
  maybeFetchPublicPlanByName,
  getBillingUpgradeUrl,
} from "@rilldata/web-admin/features/billing/plans/selectors.ts";
import {
  adminServiceRenewBillingSubscription,
  adminServiceUpdateBillingSubscription,
} from "@rilldata/web-admin/client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { triggerWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils.ts";
import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations.ts";
import { page } from "$app/stores";
import { get } from "svelte/store";

export async function upgradeToPlan(
  org: string,
  planName: string,
  categorisedIssues: CategorisedOrganizationBillingIssues,
  redirect: string | null,
) {
  if (categorisedIssues.payment.length > 0) {
    // Payment setup is required first. Carry the chosen plan through the Stripe
    // return URL so the upgrade-callback page can complete the upgrade.
    window.open(
      await fetchPaymentsPortalURL(
        org,
        getBillingUpgradeUrl(get(page), org, planName),
        categorisedIssues.needsPaymentSetup,
      ),
      "_self",
    );
    return;
  }

  const plan = await maybeFetchPublicPlanByName(planName);
  if (!plan) return;
  if (categorisedIssues.cancelled) {
    await adminServiceRenewBillingSubscription(org, {
      planName,
    });
    eventBus.emit("notification", {
      type: "success",
      message: m.billing_plan_renewed({ planName: plan.displayName }),
    });
  } else {
    await adminServiceUpdateBillingSubscription(org, {
      planName,
    });
    triggerWelcomeToRillDialog(planName);
  }
  void invalidateBillingInfo(org);
  if (redirect) {
    // redirect param could be on a different domain like the rill developer instance
    // so using goto won't work
    window.open(redirect, "_self");
  }
}
