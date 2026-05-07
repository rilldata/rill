<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceGetOrganization,
    V1BillingIssueType,
  } from "@rilldata/web-admin/client";
  import { needsPaymentSetup } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
  import {
    getBillingCycleDates,
    getBillingStatsForOrg,
    getPlanTierForSubscription,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type {
    PlanTier,
    TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/types";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CostAndUsage from "@rilldata/web-admin/features/billing/plans/modules/CostAndUsage.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import CancelPlanDialog from "@rilldata/web-admin/features/billing/plans/dialog/CancelPlanDialog.svelte";
  import UpgradeToProDialog from "@rilldata/web-admin/features/billing/plans/dialog/UpgradeToProDialog.svelte";

  let {
    organization,
    showUpgradeDialog,
    billingPortalUrl,
    cancelOpen = $bindable(false),
  }: {
    organization: string;
    showUpgradeDialog: boolean;
    billingPortalUrl: string | undefined;
    cancelOpen?: boolean;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);

  let orgQuery = $derived(createAdminServiceGetOrganization(organization));
  let hasPaymentCustomer = $derived(
    !!$orgQuery.data?.organization?.paymentCustomerId,
  );

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let paymentIssues = $derived($categorisedIssues.data?.payment);

  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);

  let currentPlan: PlanTier = $derived(
    getPlanTierForSubscription(subscription),
  );

  let isTrialExpired = $derived(
    $categorisedIssues.data?.trial?.type ===
      V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED ||
      $categorisedIssues.data?.trial?.type ===
        V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_CREDITS_DEPLETED,
  );

  let dialogType: TeamPlanDialogTypes = $derived.by(() => {
    if (subHasEnded) return "renew";
    if (isTrialExpired) return "trial-expired";
    return "base";
  });

  let renewEndDate = $derived(
    $categorisedIssues.data?.cancelled?.metadata?.subscriptionCancelled
      ?.endDate ?? "",
  );

  // Credit (Pro Trial / free plan)
  // TODO: wire usedCredit to billing usage API once available
  const TOTAL_CREDIT = 250;
  let usedCredit = $derived(0);
  let availableCredit = $derived(TOTAL_CREDIT - usedCredit);
  let creditPercent = $derived(Math.round((usedCredit / TOTAL_CREDIT) * 100));
  function fmtCredit(n: number): string {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }).format(n);
  }

  // Trial timer
  const TRIAL_DAYS = 30;
  let trialEndDate = $derived(subscription?.trialEndDate);
  let trialDaysUsed = $derived.by(() => {
    if (!trialEndDate) return 0;
    const end = new Date(trialEndDate).getTime();
    const start = end - TRIAL_DAYS * 24 * 60 * 60 * 1000;
    const now = Date.now();
    const elapsed = Math.floor((now - start) / (24 * 60 * 60 * 1000));
    return Math.max(0, Math.min(TRIAL_DAYS, elapsed));
  });
  let trialDaysRemaining = $derived(TRIAL_DAYS - trialDaysUsed);
  let trialPercent = $derived(Math.round((trialDaysUsed / TRIAL_DAYS) * 100));

  let billingStats = $derived(getBillingStatsForOrg(organization));
  let dailyRunRate = $derived(
    $billingStats.prodDailyCost + $billingStats.devDailyCost,
  );

  // Pro plan credit + post-credit estimate. Available credit is hard-zero
  // until the billing usage API exposes the remaining trial credit balance.
  let proAvailableCredit = $derived(0);
  let proEstimatedCost = $derived(
    Math.max(dailyRunRate - proAvailableCredit, 0),
  );

  // Billing cycle
  let { formattedPeriodStart, formattedPeriodEnd, formattedDueDate } = $derived(
    getBillingCycleDates(subscription),
  );

  // Upgrade dialog
  let upgradeDialogOpen = $state(false);
  $effect(() => {
    if (showUpgradeDialog) upgradeDialogOpen = true;
  });

  // Pro upgrade confirmation
  let upgradeProDialogOpen = $state(false);

  async function handleUpgradeToPro() {
    // No payment method on file, or payment issues → send to Stripe portal to set up.
    if (!hasPaymentCustomer || paymentIssues?.length) {
      const setup = paymentIssues?.length
        ? needsPaymentSetup(paymentIssues)
        : true;
      window.open(
        await fetchPaymentsPortalURL(organization, window.location.href, setup),
        "_self",
      );
      return;
    }
    // Payment method on file → confirm before upgrading.
    upgradeProDialogOpen = true;
  }

  function handleContactSales() {
    window.Pylon("show");
  }
</script>

<section>
  <h2 class="section-header">Plan</h2>

  <div class="plan-card">
    <!-- Top row: badge + description + action -->
    <div class="plan-top">
      <div class="flex items-center gap-3">
        {#if currentPlan === "enterprise"}
          <span class="plan-badge enterprise">Enterprise</span>
          <span class="plan-description">Custom contract</span>
        {:else if currentPlan === "managed"}
          <span class="plan-badge managed">Managed</span>
          <span class="plan-description">Custom contract · Fully managed</span>
        {:else if currentPlan === "trial"}
          <span class="plan-badge trial">Free Trial</span>
          {#if isTrialExpired}
            <span class="plan-description"
              >Trial expired · Projects hibernated</span
            >
          {:else}
            <span class="plan-description">30 day free trial</span>
            <Tooltip location="right" alignment="middle" distance={8}>
              <span class="text-fg-muted flex">
                <InfoCircle size="16px" />
              </span>
              <TooltipContent maxWidth="240px" slot="tooltip-content">
                Legacy free trial · 30 days, no credit card required. Projects
                hibernate when trial ends.
              </TooltipContent>
            </Tooltip>
          {/if}
        {:else if currentPlan === "free"}
          <span class="plan-badge free">Pro Trial</span>
          <span class="plan-description">$250 free credit</span>
          <Tooltip location="right" alignment="middle" distance={8}>
            <span class="text-fg-muted flex">
              <InfoCircle size="16px" />
            </span>
            <TooltipContent maxWidth="240px" slot="tooltip-content">
              No time limit, use it until it's gone.<br />$0.15/unit/hr · $1/GB
              storage/mo<br />1 unit = 4GiB RAM, 1vGPU
            </TooltipContent>
          </Tooltip>
        {:else if currentPlan === "pro"}
          <span class="plan-badge pro">Pro</span>
          <span class="plan-description">Usage based pricing</span>
          <Tooltip location="right" alignment="middle" distance={8}>
            <span class="text-fg-muted flex">
              <InfoCircle size="16px" />
            </span>
            <TooltipContent maxWidth="240px" slot="tooltip-content">
              $0.15/unit/hr · $1/GB storage/mo. Cancel anytime.
            </TooltipContent>
          </Tooltip>
        {:else if currentPlan === "team"}
          <span class="plan-badge team">Team (Legacy)</span>
          <span class="plan-description">$250/mo flat + storage</span>
          <Tooltip location="right" alignment="middle" distance={8}>
            <span class="text-fg-muted flex">
              <InfoCircle size="16px" />
            </span>
            <TooltipContent maxWidth="240px" slot="tooltip-content">
              $250/mo flat rate + storage overages. 10 GB included · $25/GB
              over. Up to 8 slots.
            </TooltipContent>
          </Tooltip>
        {/if}
      </div>

      <div class="flex items-center gap-2">
        {#if currentPlan === "enterprise"}
          <button class="contact-us-btn" onclick={handleContactSales}>
            Contact us
          </button>
        {:else if currentPlan === "trial"}
          <button
            class="subscribe-btn"
            onclick={() => (upgradeDialogOpen = true)}
          >
            Upgrade to Teams
          </button>
        {:else if currentPlan === "free"}
          <button class="subscribe-btn" onclick={handleUpgradeToPro}>
            Upgrade to Pro
          </button>
        {:else if (currentPlan === "pro" || currentPlan === "team") && billingPortalUrl}
          <a
            class="pricing-link-top"
            href={billingPortalUrl}
            target="_blank"
            rel="noreferrer noopener"
          >
            View detailed usage
            <svg
              class="w-3 h-3"
              viewBox="0 0 12 12"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path d="M3 3h6v6M3 9l6-6" />
            </svg>
          </a>
        {/if}
      </div>
    </div>

    {#if currentPlan === "enterprise" || currentPlan === "managed"}
      <p class="text-sm text-fg-tertiary mt-4 pb-4">
        Fully managed slots, dedicated CSM, white-label capabilities, and custom
        SLAs. Contact your CSM for contract details or changes.
      </p>
    {:else if currentPlan === "team"}
      <p class="text-sm text-fg-tertiary mt-4 pb-4">
        Legacy flat-rate plan. $250/mo includes up to 8 slots and 10 GB storage,
        with $25/GB for overages. Upgrade to Pro for usage-based pricing.
      </p>
    {/if}

    {#if currentPlan === "trial" && trialEndDate}
      <div class="trial-section">
        <div class="flex justify-between mb-1">
          <div>
            <span class="trial-label">Days used</span>
            <p class="trial-number-used">
              {trialDaysUsed}
            </p>
          </div>
          <div class="text-right">
            <span class="trial-label">Days remaining</span>
            <p
              class="trial-number"
              class:text-green-600={trialDaysRemaining > 7}
              class:text-red-600={trialDaysRemaining <= 7}
            >
              {trialDaysRemaining}
            </p>
          </div>
        </div>
        <div class="trial-bar-bg">
          <div class="trial-bar-fill" style:width="{trialPercent}%"></div>
        </div>
        <div class="flex justify-between mt-1">
          <span class="text-xs text-fg-tertiary">
            {trialPercent}% of trial used, projects will hibernate when trial
            ends
          </span>
          <span class="text-xs text-fg-tertiary">30 days</span>
        </div>
        <a
          class="pricing-link"
          href="https://www.rilldata.com/pricing"
          target="_blank"
          rel="noreferrer noopener"
        >
          See pricing details
          <svg
            class="w-3 h-3"
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M1 6h9M7.5 3l3 3-3 3" />
          </svg>
        </a>
      </div>
    {/if}

    {#if currentPlan === "free"}
      <div class="credit-section">
        <div class="flex justify-between">
          <span class="credit-label">Used credit</span>
          <span class="credit-label">Available credit</span>
        </div>
        <div class="flex justify-between items-end">
          <span class="credit-used">{fmtCredit(usedCredit)}</span>
          <span class="credit-available">{fmtCredit(availableCredit)}</span>
        </div>
        <div class="credit-bar-bg">
          <div class="credit-bar-fill" style:width="{creditPercent}%"></div>
        </div>
        <span class="text-xs text-fg-tertiary font-medium">
          {creditPercent}% used, projects will hibernate when credits run out.
        </span>
        <a
          class="pricing-link"
          href="https://www.rilldata.com/pricing"
          target="_blank"
          rel="noreferrer noopener"
        >
          See pricing details
          <svg
            class="w-3 h-3"
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M1 6h9M7.5 3l3 3-3 3" />
          </svg>
        </a>
      </div>
    {/if}

    {#if currentPlan === "pro"}
      <!-- TODO: replace all amounts with accrued values from billing usage API -->
      <div class="pro-stats">
        <div class="pro-stat-col">
          <span class="pro-stat-label">Current period estimate</span>
          <span class="pro-stat-amount text-fg-secondary"
            >{fmtCredit(dailyRunRate)}</span
          >
          <span class="pro-stat-sub">
            {formattedPeriodStart} – {formattedPeriodEnd}
          </span>
        </div>
        <div class="pro-stat-col border-l pl-6">
          <span class="pro-stat-label">Available credit</span>
          <span class="pro-stat-amount text-green-700"
            >{fmtCredit(proAvailableCredit)}</span
          >
          <span class="credit-pill">
            <svg viewBox="0 0 12 12" fill="none" class="w-3 h-3 shrink-0">
              <path
                d="M10 3L4.5 8.5 2 6"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            Trial credit applied to your bill
          </span>
        </div>
        <div class="pro-stat-col border-l pl-6">
          <span class="pro-stat-label">Estimated cost after applied credit</span
          >
          <span class="pro-stat-amount text-fg-primary"
            >{fmtCredit(proEstimatedCost)}</span
          >
          {#if formattedDueDate}
            <span class="pro-stat-sub">Due {formattedDueDate}</span>
          {/if}
        </div>
      </div>
    {/if}

    {#if currentPlan === "free" || currentPlan === "pro"}
      <CostAndUsage {organization} />
    {/if}
  </div>
</section>

<CancelPlanDialog bind:open={cancelOpen} {organization} />

<StartTeamPlanDialog
  bind:open={upgradeDialogOpen}
  {organization}
  type={dialogType}
  endDate={renewEndDate}
/>

<UpgradeToProDialog bind:open={upgradeProDialogOpen} {organization} />

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }

  .plan-card {
    @apply border rounded-xl bg-surface-background p-6;
    box-shadow:
      0px 1px 2px 0px rgba(0, 0, 0, 0.06),
      0px 1px 3px 0px rgba(0, 0, 0, 0.1);
  }

  .plan-top {
    @apply flex items-center justify-between;
  }

  .plan-badge {
    @apply inline-flex items-center justify-center text-xs font-semibold rounded-full border-none;
    width: 76px;
    height: 21px;
    gap: 8px;
  }

  .plan-badge.trial {
    @apply text-primary-600 bg-primary-50;
  }

  .plan-badge.pro {
    @apply text-primary-600;
    background-color: #eef2ff;
  }

  .plan-badge.team {
    @apply text-fg-secondary bg-surface-subtle;
    width: auto;
    padding: 0 8px;
  }

  .plan-badge.enterprise {
    color: #9333ea;
    background-color: #f3e8ff;
  }

  .plan-badge.managed {
    @apply text-primary-600 bg-primary-50;
  }

  .plan-badge.free {
    @apply text-fg-secondary bg-surface-subtle;
  }

  .plan-description {
    @apply font-sans font-semibold text-lg leading-7 align-middle text-fg-tertiary;
  }

  .trial-label {
    @apply font-sans font-semibold text-xs text-fg-tertiary;
    line-height: 100%;
  }

  .trial-number-used {
    @apply font-sans font-semibold text-2xl leading-8;
  }

  .trial-number {
    @apply font-sans font-semibold text-4xl leading-10;
  }

  .subscribe-btn {
    @apply text-sm font-medium text-white bg-primary-500 px-5 py-2 cursor-pointer border-none rounded-none;
  }

  .subscribe-btn:hover {
    @apply bg-primary-600;
  }

  .contact-us-btn {
    @apply text-sm font-medium text-primary-600 border border-primary-500 px-4 py-2 cursor-pointer bg-white rounded-none;
    height: 36px;
  }

  .contact-us-btn:hover {
    @apply bg-primary-50;
  }

  .credit-section {
    @apply flex flex-col gap-2 mt-4 pt-4 border-t;
  }

  .credit-label {
    @apply text-xs font-semibold text-fg-tertiary;
    line-height: 1;
  }

  .credit-used {
    @apply text-2xl font-semibold text-fg-tertiary;
  }

  .credit-available {
    @apply text-4xl font-semibold text-green-600;
  }

  .credit-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .pro-stats {
    @apply flex gap-6 mt-4 pt-4 border-t;
    min-height: 92px;
  }

  .pro-stat-col {
    @apply flex flex-col gap-2 flex-1 justify-center;
  }

  .pro-stat-label {
    @apply text-xs font-semibold text-fg-tertiary;
    line-height: 1;
  }

  .pro-stat-amount {
    @apply text-4xl font-semibold leading-none;
  }

  .pro-stat-sub {
    @apply text-sm font-medium text-fg-tertiary;
    line-height: 1;
  }

  .credit-pill {
    @apply inline-flex items-center gap-1 text-sm font-medium text-green-700 bg-green-50 rounded-full px-2.5 py-1;
  }

  .credit-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .trial-section {
    @apply mt-4 pt-4 border-t flex flex-col;
  }

  .trial-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .trial-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .pricing-link {
    @apply mt-3 inline-flex items-center gap-1 self-end;
    @apply text-xs font-medium text-primary-600 no-underline;
  }

  .pricing-link:hover {
    @apply underline;
  }

  .pricing-link-top {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-link-top:hover {
    @apply underline;
  }
</style>
