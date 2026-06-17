<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { SELF_SERVE_PLANS } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";
  import DetailedUsageLink from "@rilldata/web-admin/features/billing/plans/modules/DetailedUsageLink.svelte";

  let {
    tier,
    billingPortalUrl,
  }: {
    tier: "starter" | "growth";
    billingPortalUrl: string | undefined;
  } = $props();

  let plan = $derived(SELF_SERVE_PLANS.find((p) => p.tier === tier));
</script>

{#if plan}
  <PlanContainer
    badge={plan.displayName}
    description={`${plan.price} ${plan.priceUnit}`}
  >
    {#snippet info()}
      {plan.tagline}
    {/snippet}

    {#snippet action()}
      <DetailedUsageLink {billingPortalUrl} />
    {/snippet}

    <ul class="plan-details-container">
      {#each plan.highlights as highlight (highlight)}
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
