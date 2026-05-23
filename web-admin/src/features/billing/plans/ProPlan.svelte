<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import {
    PricingDetails,
    PricingDetailsCompact,
  } from "@rilldata/web-common/features/billing/pricing-details.ts";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge="Pro" description="Usage based pricing">
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

  {#snippet info()}
    {PricingDetailsCompact} Cancel anytime.
  {/snippet}

  <div class="text-sm text-fg-tertiary mt-4 pb-4">
    {PricingDetails}
  </div>
</PlanContainer>

<style lang="postcss">
  .pricing-link-top {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-link-top:hover {
    @apply underline;
  }
</style>
