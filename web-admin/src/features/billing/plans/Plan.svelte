<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
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

  let orgQuery = $derived(createAdminServiceGetOrganization(organization));
  let hasPaymentCustomer = $derived(
    !!$orgQuery.data?.organization?.paymentCustomerId,
  );

  let hasBillingTrialIssue = $derived(!!$categorisedIssues.data?.trial);
  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);
  let subIsProPlan = $derived(plan && isProPlan(plan.name));
  let subIsTeamPlan = $derived(plan && isTeamPlan(plan.name));
  let subIsManagedPlan = $derived(plan && isManagedPlan(plan.name));
  let subIsEnterprisePlan = $derived(
    plan && (isEnterprisePlan(plan.name) || isManagedPlan(plan.name)),
  );

  type PlanTier = "trial" | "pro" | "team" | "enterprise";
  let currentPlan: PlanTier = $derived.by(() => {
    if (subIsTeamPlan) return "team";
    if (subIsManagedPlan || subIsEnterprisePlan) return "enterprise";
    // Pro plan with payment = Pro; Pro plan without payment = Pro Trial
    if (subIsProPlan && hasPaymentCustomer) return "pro";
    // Everything else is Pro Trial: pro without payment, free_trial plan, no plan, cancelled
    return "trial";
  });

  let dialogType: TeamPlanDialogTypes = $derived.by(() => {
    if (subHasEnded) return "renew";
    if (hasBillingTrialIssue) {
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
</script>

<section>
  <h2 class="section-header">Plan</h2>

  <div class="plan-card">
    <!-- Top row: badge + description + action -->
    <div class="plan-top">
      <div class="flex items-center gap-3">
        {#if currentPlan === "enterprise"}
          <span class="plan-badge enterprise">Enterprise</span>
          <span class="text-sm text-fg-secondary"
            >Custom contract · Fully managed</span
          >
        {:else if currentPlan === "trial"}
          <span class="plan-badge trial">Pro Trial</span>
          <span class="text-sm text-fg-secondary"
            >$250 free credit · No time limit</span
          >
        {:else if currentPlan === "pro"}
          <span class="plan-badge pro">Pro</span>
          <span class="text-sm text-fg-secondary"
            >Usage based pricing · $0.15/slot/hr</span
          >
        {:else if currentPlan === "team"}
          <span class="plan-badge team">Team</span>
          <span class="text-sm text-fg-secondary">Legacy plan</span>
        {/if}
      </div>

      {#if currentPlan === "enterprise"}
        <button class="contact-btn" onclick={() => window.Pylon("show")}>
          Contact us
        </button>
      {:else if currentPlan === "trial"}
        <button class="subscribe-btn" onclick={handleSubscribe}>
          Subscribe to Pro
        </button>
      {/if}
    </div>

    {#if currentPlan === "enterprise"}
      <p class="text-sm text-fg-secondary mt-4">
        Fully managed slots, dedicated CSM, white-label capabilities, and custom
        SLAs. Contact your CSM for contract details or changes.
      </p>
    {/if}

    <!-- Divider -->
    <div class="divider"></div>

    <!-- Slots + storage row -->
    <div class="stats-row">
      <div class="flex items-center gap-4">
        <div class="stat-item">
          <span class="stat-value">{totalSlots}</span>
          <span class="stat-label">Total slots</span>
        </div>
        <span class="stat-divider"></span>
        <div class="stat-item">
          <span class="stat-value">{prodSlots}</span>
          <span class="stat-label">Prod slots</span>
        </div>
        <span class="stat-divider"></span>
        <div class="stat-item">
          <span class="stat-value">{devSlots}</span>
          <span class="stat-label">Dev slots</span>
        </div>
        <span class="stat-divider"></span>
        <div class="stat-item">
          <span class="stat-value"
            >{totalStorage > 0
              ? formatMemorySize(totalStorage)
              : "0 B"}</span
          >
          <span class="stat-label">Storage</span>
        </div>
      </div>
      <a
        href="/{organization}/-/settings/billing/usage"
        class="view-usage-link"
      >
        View usage
        <svg
          class="w-3 h-3"
          viewBox="0 0 12 12"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M4.5 2.5l4 3.5-4 3.5" />
        </svg>
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

  .subscribe-btn {
    @apply text-sm font-medium text-white bg-primary-500 px-5 py-2 cursor-pointer border-none rounded-none;
  }

  .subscribe-btn:hover {
    @apply bg-primary-600;
  }

  .contact-btn {
    @apply text-sm font-medium text-fg-primary border border-gray-300 px-5 py-2 cursor-pointer bg-transparent rounded-sm;
  }

  .contact-btn:hover {
    @apply bg-surface-subtle;
  }

  .divider {
    @apply border-t mt-4;
  }

  .stats-row {
    @apply flex items-center justify-between pt-4;
  }

  .view-usage-link {
    @apply flex items-center gap-1 text-sm font-medium text-primary-600 no-underline;
  }

  .view-usage-link:hover {
    @apply underline;
  }

  .stat-item {
    @apply flex items-center gap-1.5 text-sm text-fg-primary;
  }

  .stat-value {
    @apply font-semibold;
  }

  .stat-label {
    @apply text-fg-tertiary;
  }

  .stat-divider {
    @apply w-px h-4 bg-gray-200;
  }

  .compare-toggle {
    @apply flex items-center gap-1.5 mx-auto mt-4 text-sm font-medium text-fg-secondary cursor-pointer bg-transparent border-none;
  }

  .compare-toggle:hover {
    @apply text-fg-primary;
  }
</style>
