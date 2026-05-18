import { type CategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
import {
  fetchPaymentsPortalURL,
  fetchProPlan,
  getBillingUpgradeUrl,
} from "@rilldata/web-admin/features/billing/plans/selectors.ts";
import {
  adminServiceRenewBillingSubscription,
  adminServiceUpdateBillingSubscription,
} from "@rilldata/web-admin/client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { showWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils.ts";
import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations.ts";
import { page } from "$app/stores";
import { get } from "svelte/store";

export async function upgradeToPro(
  org: string,
  categorisedIssues: CategorisedOrganizationBillingIssues,
  redirect: string | null,
) {
  if (categorisedIssues.payment.length > 0) {
    window.open(
      await fetchPaymentsPortalURL(
        org,
        getBillingUpgradeUrl(get(page), org),
        true,
      ),
      "_self",
    );
    return;
  }

  const proPlan = await fetchProPlan();
  if (!proPlan) return;
  if (categorisedIssues.cancelled) {
    await adminServiceRenewBillingSubscription(org, {
      planName: proPlan.name,
    });
    eventBus.emit("notification", {
      type: "success",
      message: "Your Pro plan was renewed",
    });
  } else {
    await adminServiceUpdateBillingSubscription(org, {
      planName: proPlan.name,
    });
    showWelcomeToRillDialog.set(true);
  }
  void invalidateBillingInfo(org);
  if (redirect) {
    // redirect param could be on a different domain like the rill developer instance
    // so using goto won't work
    window.open(redirect, "_self");
  }
}
