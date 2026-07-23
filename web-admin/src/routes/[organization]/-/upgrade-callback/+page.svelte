<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceRenewBillingSubscription,
    createAdminServiceUpdateBillingSubscription,
    type V1BillingIssue,
  } from "@rilldata/web-admin/client";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import {
    getPaymentIssueErrorText,
    needsPaymentSetup,
  } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import {
    fetchPaymentsPortalURL,
    maybeFetchPublicPlanByName,
    getBillingUpgradeUrl,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import {
    SELF_SERVE_PLANS,
    SELF_SERVE_PLANS_BY_NAME,
  } from "@rilldata/web-admin/features/billing/plans/plan-details";
  import { triggerWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { getConfiguredRedirectOrigins } from "@rilldata/web-admin/lib/safe-redirect";
  import { createUpgradeCallbackAction } from "./upgrade-callback-action";

  export let data: PageData;
  $: ({ cancelled, paymentIssues } = data);
  $: redirect = $page.url.searchParams.get("redirect");
  // The chosen plan is carried through the Stripe return URL (see getBillingUpgradeUrl).
  // Fall back to the first self-serve plan for older links that predate the plan chooser.
  $: planName = $page.url.searchParams.get("plan") ?? SELF_SERVE_PLANS[0].name;
  $: planDetails = SELF_SERVE_PLANS_BY_NAME[planName];
  $: planDisplayName = planDetails?.displayName ?? planName;

  /**
   * Landing page to upgrade a user to team plan.
   * Is set as a return url on stripe portal.
   */
  $: organization = $page.params.organization;

  const planUpdater = createAdminServiceUpdateBillingSubscription();
  const planRenewer = createAdminServiceRenewBillingSubscription();
  let inFlight = false;
  let errorMsg = "";

  const runUpgradeCallback = createUpgradeCallbackAction<V1BillingIssue>({
    getPaymentPortalUrl: (input) =>
      fetchPaymentsPortalURL(
        input.organization,
        getBillingUpgradeUrl($page, input.organization),
        needsPaymentSetup(input.paymentIssues),
      ),
    notifyPaymentIssue: (input, portalUrl) => {
      eventBus.emit("notification", {
        type: "error",
        message: m.billing_fix_payment_issues({
          details: getPaymentIssueErrorText(input.paymentIssues),
        }),
        link: {
          text: m.billing_update_payment(),
          href: portalUrl,
        },
        options: {
          persisted: true,
        },
      });
    },
    fetchPlan: maybeFetchPublicPlanByName,
    renew: (org, selectedPlan) =>
      $planRenewer.mutateAsync({
        org,
        data: { planName: selectedPlan },
      }),
    update: (org, selectedPlan) =>
      $planUpdater.mutateAsync({
        org,
        data: { planName: selectedPlan },
      }),
    notifyRenewed: (displayName) =>
      eventBus.emit("notification", {
        type: "success",
        message: m.billing_plan_renewed({ planName: displayName }),
      }),
    showWelcome: triggerWelcomeToRillDialog,
    invalidate: invalidateBillingInfo,
    navigate: (url) => {
      void goto(url);
    },
    open: (url) => {
      window.open(url, "_self");
    },
    formatError: extractErrorMessage,
    onState: (state) => {
      inFlight = state.inFlight;
      errorMsg = state.errorMessage;
    },
  });

  onMount(() => {
    void runUpgradeCallback({
      organization,
      planName,
      cancelled,
      paymentIssues,
      redirect,
      currentUrl: $page.url.toString(),
      allowedRedirectOrigins: getConfiguredRedirectOrigins(),
    });
  });
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if inFlight}<LoadingSpinner />{/if}
    <CtaHeader variant="bold">
      {#if cancelled}
        {m.billing_renewing_plan({ plan: planDisplayName })}
      {:else}
        {m.billing_upgrading_plan({ plan: planDisplayName })}
      {/if}
    </CtaHeader>
    {#if errorMsg}
      <p class="text-md text-red-400 font-bold mb-6" role="alert">
        {errorMsg}
      </p>
    {/if}
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
