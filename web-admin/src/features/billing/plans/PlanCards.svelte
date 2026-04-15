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

  // Hide plans smaller than current. Team (legacy) shows Pro + Enterprise.
  let showTrial = $derived(currentPlan === "trial");
  let showPro = $derived(
    currentPlan === "trial" || currentPlan === "pro" || currentPlan === "team",
  );
  let showEnterprise = $derived(currentPlan !== "enterprise");

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
    "30-day free trial period",
    "Self-serve slots (2 prod slot minimum)",
    "1 GB storage included · $1/GB above",
    '"Made with Rill" badge',
  ];

  const proFeatures = [
    "Cancel anytime, data preserved 30 days",
    "Billed on prod slots + GB data overages",
    "1 GB storage included · $1/GB above",
    '"Made with Rill" badge',
  ];

  const teamFeatures = [
    "$250/mo flat charge",
    "1 GB storage included · $25/GB above",
    "Email / Chat support",
    "10 slot limit",
  ];

  const enterpriseFeatures = [
    "Custom contract pricing",
    "Fully managed slots by Rill",
    "Custom storage limits",
    "Custom colors, logo, no badge",
  ];

  let showTeam = $derived(currentPlan === "team");
</script>

<div class="compare-container">
  {#if showTrial}
    <div class="plan-col">
      <h3 class="plan-name">Free Trial</h3>
      <p class="plan-pricing-main">30 day free trial</p>
      <p class="plan-pricing-sub">No credit card required</p>

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

      <button class="col-btn current-btn" disabled>Current plan</button>
    </div>
  {/if}

  {#if showTrial && (showTeam || showPro)}
    <span class="col-divider"></span>
  {/if}

  {#if showTeam}
    <div class="plan-col">
      <h3 class="plan-name">Team plan (legacy)</h3>
      <p class="plan-pricing-main">$250/mo</p>
      <p class="plan-pricing-sub">Flat rate + storage</p>

      <ul class="feature-list">
        {#each teamFeatures as feature}
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

      <button class="col-btn current-btn" disabled>Current plan</button>
    </div>
    <span class="col-divider"></span>
  {/if}

  {#if showPro}
    <div class="plan-col">
      <h3 class="plan-name">Pro</h3>
      <p class="plan-pricing-main">Usage based pricing</p>
      <button class="estimate-cost-link" onclick={handleEstimateCost}>
        Estimate your cost →
      </button>

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
        <button class="col-btn current-btn" disabled>Current plan</button>
      {:else}
        <button class="col-btn primary-btn" onclick={handleEstimateCost}>
          Subscribe to Pro
        </button>
      {/if}
    </div>
  {/if}

  {#if showPro && showEnterprise}
    <span class="col-divider"></span>
  {/if}

  {#if showEnterprise}
    <div class="plan-col">
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

      <button class="col-btn action-btn" onclick={handleContactSales}>
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
  .compare-container {
    @apply flex bg-surface-background;
    gap: 24px;
    margin: 0 -24px -24px;
    padding: 24px 36px 36px;
    border-radius: 0 0 12px 12px;
  }

  .plan-col {
    @apply flex flex-col flex-1;
  }

  .col-divider {
    @apply w-px bg-gray-200 my-6 shrink-0;
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

  .estimate-cost-link {
    @apply font-sans font-medium text-xs leading-4 align-middle text-primary-500 cursor-pointer mt-0.5 mb-4 block bg-transparent border-none p-0 text-left;
  }
  .estimate-cost-link:hover {
    @apply text-primary-600 underline;
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

  .col-btn {
    @apply w-full py-2.5 px-4 text-sm font-medium rounded-none cursor-pointer mt-6;
  }

  .current-btn {
    @apply text-fg-tertiary bg-surface-subtle border border-gray-200 cursor-default;
  }

  .action-btn {
    @apply text-fg-primary bg-transparent border border-gray-300;
  }

  .action-btn:hover {
    @apply bg-surface-subtle;
  }

  .primary-btn {
    @apply text-white bg-primary-500 border-none;
  }

  .primary-btn:hover {
    @apply bg-primary-600;
  }
</style>
