<script lang="ts">
  import type { V1Subscription } from "@rilldata/web-admin/client";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import { Button } from "@rilldata/web-common/components/button";

  type PlanTier = "trial" | "pro" | "enterprise";

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

  const planOrder: PlanTier[] = ["trial", "pro", "enterprise"];

  function isPastPlan(plan: PlanTier): boolean {
    return planOrder.indexOf(plan) < planOrder.indexOf(currentPlan);
  }

  let upgradeDialogOpen = $state(false);
  $effect(() => {
    if (showUpgradeDialog) upgradeDialogOpen = true;
  });

  async function handleManagePlan() {
    window.open(
      await fetchPaymentsPortalURL(organization, window.location.href),
      "_self",
    );
  }

  function handleContactSales() {
    window.Pylon("show");
  }

  const trialFeatures = [
    "$250 Orb credit",
    "Self-serve slots (min enforced)",
    "$1/GB storage above 1GB",
    '"Made with Rill" badge',
  ];

  const proFeatures = [
    "Pay-as-you-go slot pricing",
    "Self-serve slots (min enforced)",
    "$1/GB storage above 1GB",
    '"Made with Rill" badge',
  ];

  const enterpriseFeatures = [
    "Custom contract pricing",
    "Fully managed slots by Rill",
    "Custom storage limits",
    "Custom colors, logo, no badge",
  ];
</script>

<section>
  <h2 class="section-header">Plan</h2>
  <div class="plan-cards-container">
    <!-- Free Trial -->
    <div
      class="plan-card"
      class:current={currentPlan === "trial"}
      class:past={isPastPlan("trial")}
    >
      <span class="current-badge" class:invisible={currentPlan !== "trial"}
        >Current plan</span
      >
      <h3 class="plan-name">Free Trial</h3>
      <p class="plan-subtitle">For individuals and small projects.</p>
      <div class="plan-pricing">
        <span class="plan-price">$0</span>
        <span class="plan-price-unit">/ month</span>
      </div>

      <!-- TODO: Credit usage bar (API not wired up yet) -->
      <div class="credit-todo">
        <p class="text-xs text-fg-disabled italic">
          Credit usage bar coming soon
        </p>
      </div>

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
    </div>

    <!-- Pro -->
    <div
      class="plan-card"
      class:current={currentPlan === "pro"}
      class:past={isPastPlan("pro")}
    >
      <span class="current-badge" class:invisible={currentPlan !== "pro"}
        >Current plan</span
      >
      <h3 class="plan-name">Pro</h3>
      <p class="plan-subtitle">For teams scaling their analytics.</p>
      <div class="plan-pricing">
        <span class="plan-price">$0.15</span>
        <span class="plan-price-unit">per slot / hour</span>
      </div>

      {#if currentPlan === "pro"}
        <Button type="secondary" wide onClick={handleManagePlan}
          >Manage plan</Button
        >
      {:else if !isPastPlan("pro")}
        <Button type="primary" wide onClick={() => (upgradeDialogOpen = true)}
          >Upgrade</Button
        >
      {/if}

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
    </div>

    <!-- Enterprise -->
    <div
      class="plan-card"
      class:current={currentPlan === "enterprise"}
      class:past={isPastPlan("enterprise")}
    >
      <span class="current-badge" class:invisible={currentPlan !== "enterprise"}
        >Current plan</span
      >
      <h3 class="plan-name">Enterprise</h3>
      <p class="plan-subtitle">For unique business needs.</p>
      <div class="plan-pricing">
        <span class="plan-price">Custom</span>
      </div>

      <button class="contact-sales-btn" onclick={handleContactSales}
        >Contact sales &rarr;</button
      >

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
    </div>
  </div>
</section>

<StartTeamPlanDialog
  bind:open={upgradeDialogOpen}
  {organization}
  type={dialogType}
  endDate={renewEndDate}
/>

<style lang="postcss">
  .plan-cards-container {
    @apply grid grid-cols-3 gap-4;
  }

  .plan-card {
    @apply relative flex flex-col border rounded-xl bg-surface-background;
    padding: 32px 26px;
  }

  .plan-card.current {
    @apply border-2 border-primary-500;
  }

  .plan-card.past {
    @apply opacity-50;
  }

  .current-badge {
    @apply inline-block self-start text-xs font-medium text-primary-600 bg-primary-50 border border-primary-500 rounded-full px-2.5 py-0.5 mb-2;
  }

  .plan-name {
    @apply text-base font-semibold text-fg-primary leading-6;
  }

  .plan-subtitle {
    @apply text-sm text-fg-tertiary mt-1 mb-3;
  }

  .plan-pricing {
    @apply flex items-baseline gap-1 mb-3;
  }

  .plan-price {
    @apply text-3xl font-bold text-fg-primary;
  }

  .plan-price-unit {
    @apply text-sm text-fg-tertiary;
  }

  .credit-todo {
    @apply mb-4 p-2 border border-dashed rounded;
  }

  .feature-list {
    @apply flex flex-col gap-2 mt-auto pt-4 border-t list-none p-0 m-0;
  }

  .feature-item {
    @apply flex items-center gap-2 text-sm text-fg-secondary;
  }

  .check-icon {
    @apply w-4 h-4 shrink-0 text-primary-500;
  }

  .plan-card.past .check-icon {
    @apply text-fg-disabled;
  }

  .contact-sales-btn {
    @apply w-full py-2 px-4 text-sm font-medium text-fg-primary border border-gray-300 rounded-md cursor-pointer bg-transparent;
  }

  .contact-sales-btn:hover {
    @apply bg-surface-subtle;
  }

  .section-header {
    @apply text-lg font-medium text-fg-primary mb-3;
  }
</style>
