<script lang="ts">
  import {
    createAdminServiceCancelBillingSubscription,
    createAdminServiceGetBillingSubscription,
    createAdminServiceListProjectsForOrganization,
    V1BillingIssueType,
    V1BillingPlanType,
  } from "@rilldata/web-admin/client";
  import { getErrorForMutation } from "@rilldata/web-admin/client/utils";
  import { invalidateBillingInfo } from "@rilldata/web-admin/features/billing/invalidations";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type {
    PlanTier,
    TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/types";
  import {
    isEnterprisePlan,
    isFreePlan,
    isManagedPlan,
    isProPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import PlanCards from "@rilldata/web-admin/features/billing/plans/PlanCards.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";

  let {
    organization,
    showUpgradeDialog,
    cancelOpen = $bindable(false),
  }: {
    organization: string;
    showUpgradeDialog: boolean;
    cancelOpen?: boolean;
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

  let currentPlan: PlanTier = $derived.by(() => {
    // Prefer planType enum when available; fall back to plan.name string matching
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_TEAM ||
      isTeamPlan(planName)
    )
      return "team";
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_MANAGED ||
      isManagedPlan(planName)
    )
      return "managed";
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
      isEnterprisePlan(planName)
    )
      return "enterprise";
    if (
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_PRO ||
      isProPlan(planName)
    )
      return "pro";
    if (isFreePlan(planName)) return "free";
    // free_trial, no plan, cancelled — all trial
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

  let usageMetrics = $derived(getOrganizationUsageMetrics(organization));
  let totalStorage = $derived(
    $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0,
  );

  // TODO: wire per-bucket accrued costs (prod compute, dev compute, storage)
  // and the current period total from the billing usage API once it exposes
  // accrued dollar amounts. Today's computed projections (slots × list rate)
  // are misleading next to real billed amounts, so the UI renders TODO
  // placeholders until the backend data is available.

  // Billing cycle
  let cycleEnd = $derived(subscription?.currentBillingCycleEndDate);
  // TODO: replace with subscription billing cycle dates once accrued cost API is available
  let periodStart = $derived.by(() => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 1).toLocaleDateString(
      undefined,
      { month: "short", day: "numeric", year: "numeric" },
    );
  });
  let periodEnd = $derived.by(() => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth() + 1, 0).toLocaleDateString(
      undefined,
      { month: "short", day: "numeric", year: "numeric" },
    );
  });
  let dueDate = $derived.by(() => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth() + 1, 1).toLocaleDateString(
      undefined,
      { month: "short", day: "numeric", year: "numeric" },
    );
  });

  // Compare plans
  let comparePlansOpen = $state(false);
  let showComparePlans = $derived(
    currentPlan !== "enterprise" && currentPlan !== "managed",
  );

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

  // Cancel subscription
  let planCanceller = $derived(createAdminServiceCancelBillingSubscription());
  let cancelError = $derived(getErrorForMutation($planCanceller));
  let cycleEndFormatted = $derived.by(() => {
    if (!cycleEnd) return "";
    return new Date(cycleEnd).toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      year: "numeric",
    });
  });
  async function handleCancelPlan() {
    await $planCanceller.mutateAsync({ org: organization });
    eventBus.emit("notification", {
      type: "success",
      message: `Your ${currentPlan === "pro" ? "Pro" : "Team"} plan was cancelled`,
    });
    void invalidateBillingInfo(organization, [
      V1BillingIssueType.BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED,
    ]);
    cancelOpen = false;
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
        {:else if currentPlan === "trial" || currentPlan === "free"}
          <button class="subscribe-btn" onclick={handleSubscribe}>
            Subscribe to Pro
          </button>
        {:else if currentPlan === "team"}
          <button class="subscribe-btn" onclick={handleSubscribe}>
            Switch to Pro
          </button>
        {/if}
      </div>
    </div>

    {#if currentPlan === "enterprise" || currentPlan === "managed"}
      <p class="text-sm text-fg-tertiary mt-4">
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
      </div>
    {/if}

    {#if currentPlan === "pro"}
      <!-- TODO: replace all amounts with accrued values from billing usage API -->
      <div class="pro-stats">
        <div class="pro-stat-col">
          <span class="pro-stat-label">Current period estimate</span>
          <span class="pro-stat-amount text-fg-secondary"
            >TODO: period total</span
          >
          <span class="pro-stat-sub">{periodStart} – {periodEnd}</span>
        </div>
        <div class="pro-stat-col border-l pl-6">
          <span class="pro-stat-label">Available credit</span>
          <span class="pro-stat-amount text-green-700"
            >TODO: available credit</span
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
            >TODO: estimated cost</span
          >
          {#if dueDate}
            <span class="pro-stat-sub">Due {dueDate}</span>
          {/if}
        </div>
      </div>
    {:else if currentPlan === "team"}
      <div class="period-estimate">
        <span class="period-label">Current period estimate</span>
        <span class="period-value">TODO: period total</span>
        <span class="period-cycle">{periodStart} – {periodEnd}</span>
      </div>
    {/if}

    {#if currentPlan !== "enterprise" && currentPlan !== "managed"}
      <!-- Cost + usage row -->
      <!-- TODO: replace per-bucket dollar values with accrued costs once the
           billing usage API exposes them. Current values (prodCost/devCost/
           storageCost) project from current config × list rate, which is
           misleading vs. actual billed amounts. -->
      <div class="stats-row">
        <div class="flex items-center gap-4">
          <div class="stat-column">
            <span class="stat-value">TODO: accrued prod cost</span>
            <span class="stat-label">{prodSlots} Prod Compute Units</span>
          </div>
          <div class="stat-column">
            <span class="stat-value">TODO: accrued dev cost</span>
            <span class="stat-label">{devSlots} Dev Compute Units</span>
          </div>
          <div class="stat-column">
            <span class="stat-value">TODO: accrued storage cost</span>
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

{#if currentPlan === "pro" || currentPlan === "team"}
  <AlertDialog bind:open={cancelOpen}>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle
          >Cancel your {currentPlan === "pro" ? "Pro" : "Team"} plan?</AlertDialogTitle
        >
        <AlertDialogDescription>
          If you cancel your plan, you'll still be able to access your account
          through
          <span class="font-semibold">{cycleEndFormatted}.</span>
        </AlertDialogDescription>
        {#if cancelError}
          <p class="text-red-500 text-sm">{cancelError}</p>
        {/if}
      </AlertDialogHeader>
      <AlertDialogFooter class="mt-3">
        <Button
          type="secondary"
          onClick={handleCancelPlan}
          loading={$planCanceller.isPending}
        >
          Cancel plan
        </Button>
        <Button type="primary" onClick={() => (cancelOpen = false)}>
          Keep plan
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
{/if}

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
