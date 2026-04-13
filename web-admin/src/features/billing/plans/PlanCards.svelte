<script lang="ts">
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";

  type PlanTier = "trial" | "pro" | "team" | "enterprise";

  let {
    organization,
    currentPlan,
    showUpgradeDialog = false,
    dialogType = "base",
    renewEndDate = "",
  }: {
    organization: string;
    currentPlan: PlanTier;
    showUpgradeDialog?: boolean;
    dialogType?: TeamPlanDialogTypes;
    renewEndDate?: string;
  } = $props();

  const planOrder: PlanTier[] = ["trial", "pro", "team", "enterprise"];

  function isPastPlan(plan: PlanTier): boolean {
    return planOrder.indexOf(plan) < planOrder.indexOf(currentPlan);
  }

  // Hide plans smaller than current. Team (legacy) shows Pro + Enterprise.
  let showTrial = $derived(currentPlan === "trial");
  let showPro = $derived(
    currentPlan === "trial" || currentPlan === "pro" || currentPlan === "team",
  );
  let showEnterprise = $derived(currentPlan !== "enterprise");

  let visibleCount = $derived(
    (showTrial ? 1 : 0) + (showPro ? 1 : 0) + (showEnterprise ? 1 : 0),
  );

  let upgradeDialogOpen = $state(false);
  $effect(() => {
    if (showUpgradeDialog) upgradeDialogOpen = true;
  });

  async function handleEstimateCost() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }

  function handleContactSales() {
    window.Pylon("show");
  }

  const trialFeatures = [
    "Credit rolls over when you subscribe to Pro",
    "Self-serve slots (2 prod slot minimum)",
    "1 GB storage included · $1/GB above",
    '"Made with Rill" badge',
  ];

  const proFeatures = [
    "Unused trial credit applied to first bill",
    "Cancel anytime, data preserved 30 days",
    "1 GB storage included · $1/GB above",
    '"Made with Rill" badge',
  ];

  const enterpriseFeatures = [
    "Custom contract pricing",
    "Fully managed slots by Rill",
    "Custom storage limits",
    "Custom colors, logo, no badge",
  ];
</script>

<div class="plan-cards-container" class:two-col={visibleCount === 2}>
  {#if showTrial}
    <!-- Pro Trial -->
    <div class="plan-card" class:current={currentPlan === "trial"}>
      <h3 class="plan-name">Pro Trial</h3>
      <p class="plan-pricing-main">$250 free credit</p>
      <p class="plan-pricing-sub">No time limit</p>

      <ul class="feature-list">
        {#each trialFeatures as feature}
          <li class="feature-item">
            <svg class="check-icon" viewBox="0 0 16 16" fill="none">
              <path
                d="M13.3 4.7a.5.5 0 0 1 0 .7l-6.5 6.5a.5.5 0 0 1-.7 0L3.2 9a.5.5 0 1 1 .7-.7l2.6 2.6 6.1-6.1a.5.5 0 0 1 .7 0Z"
                fill="currentColor"
              />
            </svg>
            <span>{feature}</span>
          </li>
        {/each}
      </ul>

      <button class="card-btn current-plan-btn" disabled>Current plan</button>
    </div>
  {/if}

  {#if showPro}
    <!-- Pro -->
    <div class="plan-card" class:current={currentPlan === "pro"}>
      <h3 class="plan-name">Pro</h3>
      <p class="plan-pricing-main">Usage based pricing</p>
      <p class="plan-pricing-sub">$0.15/slot/hr · $1/GB storage/mo</p>

      <ul class="feature-list">
        {#each proFeatures as feature}
          <li class="feature-item">
            <svg class="check-icon" viewBox="0 0 16 16" fill="none">
              <path
                d="M13.3 4.7a.5.5 0 0 1 0 .7l-6.5 6.5a.5.5 0 0 1-.7 0L3.2 9a.5.5 0 1 1 .7-.7l2.6 2.6 6.1-6.1a.5.5 0 0 1 .7 0Z"
                fill="currentColor"
              />
            </svg>
            <span>{feature}</span>
          </li>
        {/each}
      </ul>

      {#if currentPlan === "pro"}
        <button class="card-btn current-plan-btn" disabled>Current plan</button>
      {:else}
        <button class="card-btn action-btn" onclick={handleEstimateCost}>
          Estimate your cost
        </button>
      {/if}
    </div>
  {/if}

  {#if showEnterprise}
    <!-- Enterprise -->
    <div class="plan-card">
      <h3 class="plan-name">Enterprise</h3>
      <p class="plan-pricing-main">Custom pricing</p>
      <p class="plan-pricing-sub">Annual contract</p>

      <ul class="feature-list">
        {#each enterpriseFeatures as feature}
          <li class="feature-item">
            <svg class="check-icon" viewBox="0 0 16 16" fill="none">
              <path
                d="M13.3 4.7a.5.5 0 0 1 0 .7l-6.5 6.5a.5.5 0 0 1-.7 0L3.2 9a.5.5 0 1 1 .7-.7l2.6 2.6 6.1-6.1a.5.5 0 0 1 .7 0Z"
                fill="currentColor"
              />
            </svg>
            <span>{feature}</span>
          </li>
        {/each}
      </ul>

      <button class="card-btn action-btn" onclick={handleContactSales}>
        Contact sales
      </button>
    </div>
  {/if}
</div>

<StartTeamPlanDialog
  bind:open={upgradeDialogOpen}
  {organization}
  type={dialogType}
  endDate={renewEndDate}
/>

<style lang="postcss">
  .plan-cards-container {
    @apply grid grid-cols-3 gap-4 mt-4;
  }

  .plan-cards-container.two-col {
    @apply grid-cols-2;
  }

  .plan-card {
    @apply relative flex flex-col border rounded-xl bg-surface-background;
    padding: 32px 26px;
    box-shadow:
      0px 1px 2px 0px rgba(0, 0, 0, 0.06),
      0px 1px 3px 0px rgba(0, 0, 0, 0.1);
  }

  .plan-card.current {
    @apply border-2 border-primary-500;
  }

  .plan-name {
    @apply text-sm font-medium text-fg-secondary mb-1;
  }

  .plan-pricing-main {
    @apply text-lg font-bold text-fg-primary;
  }

  .plan-pricing-sub {
    @apply text-sm text-fg-tertiary mt-0.5 mb-4;
  }

  .feature-list {
    @apply flex flex-col gap-2.5 pt-4 border-t list-none p-0 m-0 flex-1;
  }

  .feature-item {
    @apply flex items-start gap-2 text-sm text-fg-secondary;
  }

  .check-icon {
    @apply w-4 h-4 shrink-0 text-primary-500 mt-0.5;
  }

  .card-btn {
    @apply w-full py-2.5 px-4 text-sm font-medium rounded-md cursor-pointer mt-6;
  }

  .current-plan-btn {
    @apply text-fg-tertiary bg-surface-subtle border border-gray-200 cursor-default;
  }

  .action-btn {
    @apply text-fg-primary bg-transparent border border-gray-300;
  }

  .action-btn:hover {
    @apply bg-surface-subtle;
  }
</style>
