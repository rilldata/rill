import { getSafeRedirectTarget } from "@rilldata/web-admin/lib/safe-redirect.ts";

type UpgradePlan = {
  displayName?: string | null;
};

export type UpgradeCallbackInput<PaymentIssue> = {
  organization: string;
  planName: string;
  cancelled: boolean;
  paymentIssues: PaymentIssue[];
  redirect: string | null;
  currentUrl: string;
  allowedRedirectOrigins: readonly string[];
};

export type UpgradeCallbackState = {
  inFlight: boolean;
  completed: boolean;
  errorMessage: string;
};

export type UpgradeCallbackDependencies<PaymentIssue> = {
  getPaymentPortalUrl: (
    input: UpgradeCallbackInput<PaymentIssue>,
  ) => Promise<string>;
  notifyPaymentIssue: (
    input: UpgradeCallbackInput<PaymentIssue>,
    portalUrl: string,
  ) => void;
  fetchPlan: (planName: string) => Promise<UpgradePlan | null | undefined>;
  renew: (organization: string, planName: string) => Promise<unknown>;
  update: (organization: string, planName: string) => Promise<unknown>;
  notifyRenewed: (displayName: string) => void;
  showWelcome: (planName: string) => void;
  invalidate: (organization: string) => unknown;
  navigate: (url: string) => void;
  open: (url: string) => void;
  formatError: (error: unknown) => string;
  onState: (state: UpgradeCallbackState) => void;
};

async function performUpgradeCallback<PaymentIssue>(
  input: UpgradeCallbackInput<PaymentIssue>,
  dependencies: UpgradeCallbackDependencies<PaymentIssue>,
) {
  if (input.paymentIssues.length) {
    const portalUrl = await dependencies.getPaymentPortalUrl(input);
    if (!portalUrl) throw new Error("Payment portal URL is unavailable");
    dependencies.notifyPaymentIssue(input, portalUrl);
    dependencies.navigate(`/${input.organization}/-/settings/billing`);
    return;
  }

  const plan = await dependencies.fetchPlan(input.planName);
  if (!plan) throw new Error(`Plan ${input.planName} not found`);

  const safeRedirect = getSafeRedirectTarget(input.redirect, input.currentUrl, {
    allowedOrigins: input.allowedRedirectOrigins,
  });

  if (input.cancelled) {
    await dependencies.renew(input.organization, input.planName);
    dependencies.notifyRenewed(plan.displayName ?? "");
  } else {
    await dependencies.update(input.organization, input.planName);
    if (!safeRedirect) dependencies.showWelcome(input.planName);
  }

  // Invalidation and navigation are post-mutation effects. They must never run
  // when the billing request rejects.
  void dependencies.invalidate(input.organization);
  if (safeRedirect) {
    dependencies.open(safeRedirect);
  } else {
    dependencies.navigate(`/${input.organization}`);
  }
}

export function createUpgradeCallbackAction<PaymentIssue>(
  dependencies: UpgradeCallbackDependencies<PaymentIssue>,
) {
  let inFlight: Promise<void> | null = null;
  let completed = false;

  return function runUpgradeCallback(
    input: UpgradeCallbackInput<PaymentIssue>,
  ): Promise<void> {
    // onMount or repeated UI events may race; only the active request owns the
    // mutation and its redirect.
    if (inFlight !== null) return inFlight;
    if (completed) return Promise.resolve();

    dependencies.onState({
      inFlight: true,
      completed: false,
      errorMessage: "",
    });

    const request = performUpgradeCallback(input, dependencies)
      .then(() => {
        completed = true;
        dependencies.onState({
          inFlight: false,
          completed: true,
          errorMessage: "",
        });
      })
      .catch((error) => {
        dependencies.onState({
          inFlight: false,
          completed: false,
          errorMessage: dependencies.formatError(error),
        });
      })
      .finally(() => {
        // Failures clear the guard so an explicit retry can issue a fresh
        // request; successful callbacks remain completed and cannot duplicate.
        if (inFlight === request) inFlight = null;
      });

    inFlight = request;
    return request;
  };
}
