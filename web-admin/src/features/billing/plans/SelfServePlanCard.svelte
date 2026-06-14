<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { SELF_SERVE_PLANS } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";

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
      {#if billingPortalUrl}
        <a
          class="pricing-link-top"
          href={billingPortalUrl}
          target="_blank"
          rel="noreferrer noopener"
        >
          View detailed usage
          <svg
            class="w-3 h-3"
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M3 3h6v6M3 9l6-6" />
          </svg>
        </a>
      {/if}
    {/snippet}

    <ul class="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1.5 mt-4 pb-4">
      {#each plan.highlights as highlight (highlight)}
        <li class="text-sm text-fg-tertiary">{highlight}</li>
      {/each}
    </ul>
  </PlanContainer>
{/if}

<style lang="postcss">
  .pricing-link-top {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-link-top:hover {
    @apply underline;
  }
</style>
