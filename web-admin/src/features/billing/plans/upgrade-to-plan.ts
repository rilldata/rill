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
import {
  getConfiguredRedirectOrigins,
  getSafeRedirectTarget,
} from "@rilldata/web-admin/lib/safe-redirect.ts";

type UpgradePlan = {
  displayName?: string | null;
};

export type UpgradeToPlanDependencies = {
  getPage: () => Parameters<typeof getBillingUpgradeUrl>[0];
  getCurrentUrl: () => string;
  getAllowedRedirectOrigins: () => readonly string[];
  fetchPaymentPortalUrl: (
    org: string,
    returnUrl: string,
    setup: boolean,
  ) => Promise<string>;
  fetchPlan: (planName: string) => Promise<UpgradePlan | null | undefined>;
  renew: (org: string, planName: string) => Promise<unknown>;
  update: (org: string, planName: string) => Promise<unknown>;
  notifyRenewed: (displayName: string) => void;
  showWelcome: (planName: string) => void;
  invalidate: (org: string) => unknown;
  open: (url: string) => void;
};

const defaultDependencies: UpgradeToPlanDependencies = {
  getPage: () => get(page),
  getCurrentUrl: () => window.location.href,
  getAllowedRedirectOrigins: getConfiguredRedirectOrigins,
  fetchPaymentPortalUrl: fetchPaymentsPortalURL,
  fetchPlan: maybeFetchPublicPlanByName,
  renew: (org, planName) =>
    adminServiceRenewBillingSubscription(org, { planName }),
  update: (org, planName) =>
    adminServiceUpdateBillingSubscription(org, { planName }),
  notifyRenewed: (displayName) =>
    eventBus.emit("notification", {
      type: "success",
      message: m.billing_plan_renewed({ planName: displayName }),
    }),
  showWelcome: triggerWelcomeToRillDialog,
  invalidate: invalidateBillingInfo,
  open: (url) => {
    window.open(url, "_self");
  },
};

async function performUpgradeToPlan(
  dependencies: UpgradeToPlanDependencies,
  org: string,
  planName: string,
  categorisedIssues: CategorisedOrganizationBillingIssues,
  redirect: string | null,
) {
  if (categorisedIssues.payment.length > 0) {
    // The server-issued portal URL is short-lived and signed, so only the
    // winning in-flight request may fetch and consume it.
    const portalUrl = await dependencies.fetchPaymentPortalUrl(
      org,
      getBillingUpgradeUrl(dependencies.getPage(), org, planName),
      categorisedIssues.needsPaymentSetup,
    );
    if (!portalUrl) throw new Error("Payment portal URL is unavailable");
    dependencies.open(portalUrl);
    return;
  }

  const plan = await dependencies.fetchPlan(planName);
  if (!plan) throw new Error(`Plan ${planName} not found`);

  const safeRedirect = getSafeRedirectTarget(
    redirect,
    dependencies.getCurrentUrl(),
    { allowedOrigins: dependencies.getAllowedRedirectOrigins() },
  );

  if (categorisedIssues.cancelled) {
    await dependencies.renew(org, planName);
    dependencies.notifyRenewed(plan.displayName ?? "");
  } else {
    await dependencies.update(org, planName);
    // A rejected redirect behaves like no redirect, so the successful upgrade
    // still gets its normal welcome state.
    if (!safeRedirect) dependencies.showWelcome(planName);
  }

  // Mutation success is the commit point. Failed requests never invalidate
  // cached billing data or navigate away from the error shown by the dialog.
  void dependencies.invalidate(org);
  if (safeRedirect) dependencies.open(safeRedirect);
}

export function createUpgradeToPlanAction(
  dependencies: UpgradeToPlanDependencies,
) {
  const inFlight = new Map<string, Promise<void>>();

  return function runUpgradeToPlan(
    org: string,
    planName: string,
    categorisedIssues: CategorisedOrganizationBillingIssues,
    redirect: string | null,
  ): Promise<void> {
    const key = JSON.stringify([org, planName]);
    const activeRequest = inFlight.get(key);
    if (activeRequest !== undefined) return activeRequest;

    const request = performUpgradeToPlan(
      dependencies,
      org,
      planName,
      categorisedIssues,
      redirect,
    ).finally(() => {
      // Release both successes and failures; a failed attempt can be retried,
      // while rapid repeated clicks still share exactly one mutation.
      if (inFlight.get(key) === request) inFlight.delete(key);
    });
    inFlight.set(key, request);
    return request;
  };
}

export const upgradeToPlan = createUpgradeToPlanAction(defaultDependencies);
