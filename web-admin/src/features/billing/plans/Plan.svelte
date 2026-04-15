<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListProjectsForOrganization,
    V1BillingPlanType,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import {
    isEnterprisePlan,
    isManagedPlan,
    isProPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import PlanCards from "@rilldata/web-admin/features/billing/plans/PlanCards.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";

  let {
    organization,
    showUpgradeDialog,
  }: {
    organization: string;
    showUpgradeDialog: boolean;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);
  let plan = $derived(subscription?.plan);

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);
  let planType = $derived(plan?.planType);
  let planName = $derived(plan?.name ?? "");

  type PlanTier = "trial" | "pro" | "team" | "enterprise";
  let currentPlan: PlanTier = $derived.by(() => {
    // Prefer planType enum when available; fall back to plan.name string matching
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_TEAM ||
      isTeamPlan(planName)
    )
      return "team";
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_MANAGED ||
      isManagedPlan(planName) ||
      isEnterprisePlan(planName)
    )
      return "enterprise";
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_PRO ||
      isProPlan(planName)
    )
      return "pro";
    // free_trial, free, no plan, cancelled — all trial
    return "trial";
  });

  let isTrialExpired = $derived(
    $categorisedIssues.data?.trial?.type === "BILLING_ISSUE_TYPE_TRIAL_ENDED",
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

  // Slots + storage data
  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization),
  );
  let projects = $derived($projectsQuery.data?.projects ?? []);

  let prodSlots = $derived(
    projects.reduce((sum, p) => sum + Number(p.prodSlots ?? 0), 0),
  );
  let devSlots = $derived(
    projects.reduce((sum, p) => sum + Number(p.devSlots ?? 0), 0),
  );
  let totalSlots = $derived(prodSlots + devSlots);

  let usageMetrics = $derived(getOrganizationUsageMetrics(organization));
  let totalStorage = $derived(
    $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0,
  );

  // Est. monthly cost: $0.15/slot/hr × 730 hrs/mo
  const fmtCost = (slots: number) =>
    (slots * 0.15 * 730).toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    });
  let prodCost = $derived(fmtCost(prodSlots));
  let devCost = $derived(fmtCost(devSlots));

  // Storage cost: $1/GB/mo
  let storageCost = $derived(
    (totalStorage / 1e9).toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }),
  );

  // Total period estimate
  let totalEstimate = $derived(
    (totalSlots * 0.15 * 730 + totalStorage / 1e9).toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }),
  );

  // Billing cycle
  let cycleStart = $derived(subscription?.currentBillingCycleStartDate);
  let cycleEnd = $derived(subscription?.currentBillingCycleEndDate);
  function formatCycleDate(dateStr: string | undefined): string {
    if (!dateStr) return "";
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  // Compare plans
  let comparePlansOpen = $state(false);
  let showComparePlans = $derived(currentPlan !== "enterprise");

  // Upgrade dialog
  let upgradeDialogOpen = $state(false);
  $effect(() => {
    if (showUpgradeDialog) upgradeDialogOpen = true;
  });

  async function handleSubscribe() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
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
          <span class="plan-description">Custom contract · Fully managed</span>
        {:else if currentPlan === "trial"}
          <span class="plan-badge trial">Free Trial</span>
          {#if isTrialExpired}
            <span class="plan-description"
              >Trial expired · Projects hibernated</span
            >
          {:else}
            <span class="plan-description">30 day free trial</span>
          {/if}
        {:else if currentPlan === "pro"}
          <span class="plan-badge pro">Pro</span>
          <span class="plan-description"
            >Usage based pricing. $0.15/unit/hr · $1/GB storage/mo</span
          >
        {:else if currentPlan === "team"}
          <span class="plan-badge team">Team (Legacy)</span>
          <span class="plan-description">$250/mo flat + storage</span>
        {/if}
      </div>

      <div class="flex items-center gap-2">
        {#if currentPlan === "trial"}
          <button class="subscribe-btn" onclick={handleSubscribe}>
            Subscribe to Pro
          </button>
        {:else if currentPlan === "team"}
          <button class="subscribe-btn" onclick={handleSubscribe}>
            Switch to Pro
          </button>
          <button class="contact-btn" onclick={handleContactSales}>
            Upgrade to Enterprise
          </button>
        {/if}
        <Tooltip location="left" alignment="middle" distance={8}>
          <span class="text-fg-muted flex">
            <InfoCircle size="16px" />
          </span>
          <TooltipContent maxWidth="240px" slot="tooltip-content">
            $0.15/unit/hr · $1/GB storage/mo. Cancel anytime.
          </TooltipContent>
        </Tooltip>
      </div>
    </div>

    {#if currentPlan === "enterprise"}
      <p class="text-sm text-fg-secondary mt-4">
        Fully managed slots, dedicated CSM, white-label capabilities, and custom
        SLAs. Contact your CSM for contract details or changes.
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
      </div>
    {/if}

    {#if currentPlan !== "enterprise" && currentPlan !== "trial"}
      <div class="period-estimate">
        <span class="period-label">Current period estimate</span>
        <span class="period-value">{totalEstimate}</span>
        {#if cycleStart || cycleEnd}
          <span class="period-cycle"
            >{formatCycleDate(cycleStart)} – {formatCycleDate(cycleEnd)}</span
          >
        {/if}
      </div>
    {/if}

    {#if currentPlan !== "enterprise"}
      <!-- Cost + usage row -->
      <div class="stats-row">
        <div class="flex items-center gap-4">
          <div class="stat-column">
            <span class="stat-value">{prodCost}</span>
            <span class="stat-label">{prodSlots} Prod Compute Units</span>
          </div>
          <div class="stat-column">
            <span class="stat-value">{devCost}</span>
            <span class="stat-label">{devSlots} Dev Compute Units</span>
          </div>
          <div class="stat-column">
            <span class="stat-value">{storageCost}</span>
            <span class="stat-label"
              >{totalStorage > 0 ? formatMemorySize(totalStorage) : "0 B"} Storage</span
            >
          </div>
        </div>
        <a
          href="/{organization}/-/settings/billing/usage"
          class="view-usage-link"
        >
          View Usage
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

    <!-- Compare plans toggle -->
    {#if showComparePlans}
      <button
        class="compare-toggle"
        class:open={comparePlansOpen}
        onclick={() => (comparePlansOpen = !comparePlansOpen)}
      >
        Compare plans
        <svg
          class="w-4 h-4 transition-transform"
          class:rotate-180={!comparePlansOpen}
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M4 10l4-4 4 4" />
        </svg>
      </button>

      {#if comparePlansOpen}
        <PlanCards
          {organization}
          {currentPlan}
          {showUpgradeDialog}
          {dialogType}
          {renewEndDate}
        />
      {/if}
    {/if}
  </div>
</section>

<StartTeamPlanDialog
  bind:open={upgradeDialogOpen}
  {organization}
  type={dialogType}
  endDate={renewEndDate}
/>

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
    @apply text-primary-600 bg-primary-50;
  }

  .plan-badge.team {
    @apply text-fg-secondary bg-surface-subtle;
  }

  .plan-badge.enterprise {
    @apply text-primary-600 bg-primary-50;
  }

  .plan-description {
    @apply font-sans font-semibold text-lg leading-7 align-middle text-fg-tertiary;
  }

  .estimate-cost-link {
    @apply font-sans font-semibold text-lg leading-7 align-middle text-primary-500 no-underline cursor-pointer;
  }
  .estimate-cost-link:hover {
    @apply text-primary-600 underline;
  }

  .period-estimate {
    @apply flex flex-col items-center mt-4 pt-6 pb-4;
    gap: 8px;
  }

  .period-label {
    @apply font-sans font-semibold text-xs text-fg-tertiary;
    line-height: 100%;
  }

  .period-value {
    @apply font-sans font-semibold text-4xl leading-10 text-fg-primary;
  }

  .period-cycle {
    @apply font-sans font-medium text-xs text-fg-tertiary;
    line-height: 100%;
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

  .contact-btn {
    @apply text-sm font-medium text-fg-primary border border-gray-300 px-5 py-2 cursor-pointer bg-transparent rounded-none;
  }

  .contact-btn:hover {
    @apply bg-surface-subtle;
  }

  .trial-section {
    @apply mt-4 pt-4 border-t;
  }

  .trial-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .trial-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .stats-row {
    @apply flex items-center justify-between bg-surface-subtle border-t;
    margin: 16px -24px 0;
    padding: 12px 24px;
  }

  .view-usage-link {
    @apply flex items-center gap-1 text-sm font-medium text-primary-600 no-underline;
  }

  .view-usage-link:hover {
    @apply underline;
  }

  .stat-column {
    @apply flex flex-col gap-1;
  }

  .stat-value {
    @apply font-sans font-medium text-sm;
    line-height: 100%;
  }

  .stat-label {
    @apply font-sans font-medium text-xs text-fg-tertiary;
    line-height: 100%;
  }

  .stat-divider {
    @apply w-px h-4 bg-gray-200;
  }

  .compare-toggle {
    @apply flex items-center justify-center gap-1.5 text-sm font-medium text-fg-secondary cursor-pointer bg-transparent border-t border-l-0 border-r-0 border-b-0;
    margin: 0 -24px -24px;
    width: calc(100% + 48px);
    padding: 12px 0;
    border-radius: 0 0 12px 12px;
  }

  .compare-toggle.open {
    margin-bottom: 0;
    border-radius: 0;
  }

  .compare-toggle:hover {
    @apply text-fg-primary;
  }
</style>
