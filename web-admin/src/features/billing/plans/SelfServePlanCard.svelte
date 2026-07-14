<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import {
    resolvePlanHighlights,
    SELF_SERVE_PLANS,
    getTranslatedPlanDisplayName,
    getTranslatedPlanTagline,
    getTranslatedPlanPriceUnit,
  } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";
  import type { V1BillingPlan } from "@rilldata/web-admin/client";

  let {
    plan,
    billingPortalUrl,
  }: {
    plan: V1BillingPlan;
    billingPortalUrl: string | undefined;
  } = $props();

  let planDetails = $derived(
    SELF_SERVE_PLANS.find((p) => p.name === plan.name),
  );
  let highlights = $derived(
    resolvePlanHighlights(planDetails, plan.quotas ?? {}),
  );
</script>

{#if planDetails}
  <PlanContainer
    badge={getTranslatedPlanDisplayName(planDetails.name)}
    description={`${planDetails.price} ${getTranslatedPlanPriceUnit()}`}
  >
    {#snippet info()}
      {getTranslatedPlanTagline(planDetails.name)}
    {/snippet}

    {#snippet action()}
      <DetailedUsageLink {billingPortalUrl} />
    {/snippet}

    <ul class="plan-details-container">
      {#each highlights as highlight (highlight)}
        <li class="plan-details-item">{highlight}</li>
      {/each}
    </ul>
  </PlanContainer>
{/if}

<style lang="postcss">
  .plan-details-container {
    @apply grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1.5 mt-4 py-4 border-t;
  }

  .plan-details-item {
    @apply text-sm text-fg-tertiary;
  }
</style>
