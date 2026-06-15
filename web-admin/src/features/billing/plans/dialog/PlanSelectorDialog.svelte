<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { type V1BillingPlan } from "@rilldata/web-admin/client";
  import {
    isFreePlan,
    isTeamPlan,
    isTrialPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import {
    CreditsTrialPlan,
    GrowthPaidPlan,
    LegacyTeamPlan,
    LegacyTrialPlan,
    StarterPaidPlan,
    EnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/plans.ts";
  import { Button } from "@rilldata/web-common/components/button";

  let {
    current,
  }: {
    current: V1BillingPlan;
  } = $props();

  let isLegacyTrialPlan = $derived(isTrialPlan(current.name));
  let isFreeTrailPlan = $derived(isFreePlan(current.name));

  let isLegacyTeamPlan = $derived(isTeamPlan(current.name));
  let Plans = $derived([
    ...(isLegacyTrialPlan ? [LegacyTrialPlan] : []),
    ...(isFreeTrailPlan ? [CreditsTrialPlan] : []),
    ...(isLegacyTeamPlan
      ? [LegacyTeamPlan]
      : [StarterPaidPlan, GrowthPaidPlan]),
    EnterprisePlan,
  ]);

  function selectPlan(planName: string) {
    console.log(planName);
  }
</script>

<Dialog.Root open>
  <Dialog.Content class="min-w-[800px]">
    <div class="plans-container">
      {#each Plans as plan (plan.name)}
        {@const isCurrent = plan.name === current.name}

        <span class="col-divider"></span>

        <div class="plan-col">
          <h3 class="plan-name">{plan.title}</h3>
          <p class="plan-pricing-main">{plan.main}</p>
          <p class="plan-pricing-sub">{plan.sub}</p>

          <ul class="feature-list">
            {#each plan.features as feature, i (i)}
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

          {#if plan.custom}
            <Button type="primary">Contact sales</Button>
          {:else if plan.paid || isCurrent}
            <Button
              disabled={isCurrent}
              type={plan.paid ? "primary" : "secondary"}
              onClick={() => selectPlan(plan.name)}
            >
              {isCurrent ? "Current" : "Select"}
            </Button>
          {/if}
        </div>
      {/each}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .plans-container {
    @apply flex gap-6 border rounded-xl p-6 bg-surface-background;
  }

  .plan-col {
    @apply flex flex-col flex-1;
  }

  .col-divider {
    @apply w-px bg-gray-200 my-6 shrink-0;
  }
  .col-divider:nth-child(1) {
    @apply hidden;
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
    @apply flex flex-col gap-2.5 pt-4 border-t list-none p-0 m-2 flex-1;
  }

  .feature-item {
    @apply flex items-start gap-2 text-sm text-fg-secondary;
  }

  .check-icon {
    @apply w-4 h-4 shrink-0 text-primary-500 mt-0.5;
  }
</style>
