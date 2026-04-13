<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import PlanCards from "@rilldata/web-admin/features/billing/plans/PlanCards.svelte";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import {
    isEnterprisePlan,
    isManagedPlan,
    isTeamPlan,
    isFreePlan,
    isTrialPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

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

  let neverSubbed = $derived(!!$categorisedIssues.data?.neverSubscribed);
  let isTrial = $derived(!!$categorisedIssues.data?.trial);
  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);
  let subIsTeamPlan = $derived(plan && isTeamPlan(plan.name));
  let subIsManagedPlan = $derived(plan && isManagedPlan(plan.name));
  let subIsEnterprisePlan = $derived(
    plan && (isEnterprisePlan(plan.name) || isManagedPlan(plan.name)),
  );

  type PlanTier = "trial" | "pro" | "team" | "enterprise";
  let currentPlan: PlanTier = $derived.by(() => {
    if (neverSubbed || isTrial || subHasEnded) return "trial";
    if (subIsTeamPlan) return "team";
    if (subIsManagedPlan || subIsEnterprisePlan) return "enterprise";
    return "pro";
  });

  let dialogType: TeamPlanDialogTypes = $derived.by(() => {
    if (subHasEnded) return "renew";
    if (isTrial) {
      const trialIssue = $categorisedIssues.data?.trial;
      if (trialIssue?.type === "BILLING_ISSUE_TYPE_TRIAL_ENDED")
        return "trial-expired";
      return "base";
    }
    return "base";
  });

  let renewEndDate = $derived(
    $categorisedIssues.data?.cancelled?.metadata?.subscriptionCancelled
      ?.endDate ?? "",
  );

  let comparePlansOpen = $state(false);

  // Enterprise doesn't show compare plans at all
  let showComparePlans = $derived(currentPlan !== "enterprise");
</script>

<section>
  <h2 class="section-header">Plan</h2>

  <!-- Plan summary card -->
  <div class="plan-summary-card">
    {#if currentPlan === "enterprise"}
      <!-- Enterprise summary -->
      <div class="summary-top">
        <div class="flex items-center gap-3">
          <span class="plan-badge enterprise">Enterprise</span>
          <span class="text-sm text-fg-secondary"
            >Custom contract · Fully managed</span
          >
        </div>
        <button class="contact-btn" onclick={() => window.Pylon("show")}>
          Contact us
        </button>
      </div>
      <p class="text-sm text-fg-secondary mt-4">
        Fully managed slots, dedicated CSM, white-label capabilities, and custom
        SLAs. Contact your CSM for contract details or changes.
      </p>
    {:else if currentPlan === "trial"}
      <!-- Trial summary -->
      <div class="summary-top">
        <div class="flex items-center gap-3">
          <span class="plan-badge trial">Pro Trial</span>
          <span class="text-sm text-fg-secondary"
            >$250 free credit · No time limit</span
          >
        </div>
        <button class="subscribe-btn" onclick={() => (comparePlansOpen = true)}>
          Subscribe to Pro
        </button>
      </div>

      <!-- TODO: Credit usage (needs API) -->
      <div class="credit-section">
        <div class="flex justify-between mb-1">
          <div>
            <span class="text-xs text-fg-tertiary">Used credit</span>
            <p class="text-2xl font-light text-fg-secondary">—</p>
          </div>
          <div class="text-right">
            <span class="text-xs text-fg-tertiary">Available credit</span>
            <p class="text-2xl font-light text-green-600">—</p>
          </div>
        </div>
        <div class="credit-bar-bg">
          <div class="credit-bar-fill" style:width="0%"></div>
        </div>
        <div class="flex justify-between mt-1">
          <span class="text-xs text-fg-tertiary">Credit usage coming soon</span>
          <span class="text-xs text-fg-tertiary">$0.15/slot/hr</span>
        </div>
      </div>
    {:else if currentPlan === "pro"}
      <!-- Pro summary -->
      <div class="summary-top">
        <div class="flex items-center gap-3">
          <span class="plan-badge pro">Pro</span>
          <span class="text-sm text-fg-secondary"
            >Usage based pricing · $0.15/slot/hr</span
          >
        </div>
      </div>
    {:else if currentPlan === "team"}
      <!-- Team (Legacy) summary -->
      <div class="summary-top">
        <div class="flex items-center gap-3">
          <span class="plan-badge team">Team</span>
          <span class="text-sm text-fg-secondary">Legacy plan</span>
        </div>
      </div>
    {/if}

    <!-- Slots bar (always shown) -->
    <div class="slots-bar">
      <div class="flex items-center gap-4">
        <span class="slot-item"
          ><strong>—</strong> <span class="slot-label">Total slots</span></span
        >
        <span class="slot-divider"></span>
        <span class="slot-item"
          ><strong>—</strong> <span class="slot-label">Prod slots</span></span
        >
        <span class="slot-divider"></span>
        <span class="slot-item"
          ><strong>—</strong> <span class="slot-label">Dev slots</span></span
        >
        <span class="slot-divider"></span>
        <span class="slot-item"
          ><strong>—</strong> <span class="slot-label">Storage</span></span
        >
      </div>
      <a
        href="/{organization}/-/settings/billing/usage"
        class="view-usage-link"
      >
        View usage →
      </a>
    </div>
  </div>

  <!-- Compare plans toggle -->
  {#if showComparePlans}
    <button
      class="compare-toggle"
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
</section>

<style lang="postcss">
  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }

  .plan-summary-card {
    @apply border rounded-xl bg-surface-background p-6;
    box-shadow:
      0px 1px 2px 0px rgba(0, 0, 0, 0.06),
      0px 1px 3px 0px rgba(0, 0, 0, 0.1);
  }

  .summary-top {
    @apply flex items-center justify-between;
  }

  .plan-badge {
    @apply inline-block text-xs font-semibold rounded-full px-3 py-1 border;
  }

  .plan-badge.trial {
    @apply text-primary-600 bg-primary-50 border-primary-500;
  }

  .plan-badge.pro {
    @apply text-primary-600 bg-primary-50 border-primary-500;
  }

  .plan-badge.team {
    @apply text-fg-secondary bg-surface-subtle border-gray-300;
  }

  .plan-badge.enterprise {
    @apply text-primary-600 bg-primary-50 border-primary-500;
  }

  .subscribe-btn {
    @apply text-sm font-medium text-white bg-green-600 rounded-full px-5 py-2 cursor-pointer border-none;
  }

  .subscribe-btn:hover {
    @apply bg-green-700;
  }

  .contact-btn {
    @apply text-sm font-medium text-fg-primary border border-gray-300 rounded-full px-5 py-2 cursor-pointer bg-transparent;
  }

  .contact-btn:hover {
    @apply bg-surface-subtle;
  }

  .credit-section {
    @apply mt-4 pt-4 border-t;
  }

  .credit-bar-bg {
    @apply w-full h-2 bg-gray-200 rounded-full overflow-hidden;
  }

  .credit-bar-fill {
    @apply h-full bg-primary-500 rounded-full transition-all;
  }

  .slots-bar {
    @apply flex items-center justify-between mt-4 pt-4 border-t;
  }

  .slot-item {
    @apply text-sm text-fg-primary;
  }

  .slot-label {
    @apply text-fg-tertiary;
  }

  .slot-divider {
    @apply w-px h-4 bg-gray-200;
  }

  .view-usage-link {
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .view-usage-link:hover {
    @apply underline;
  }

  .compare-toggle {
    @apply flex items-center gap-1.5 mx-auto mt-4 text-sm font-medium text-fg-secondary cursor-pointer bg-transparent border-none;
  }

  .compare-toggle:hover {
    @apply text-fg-primary;
  }
</style>
