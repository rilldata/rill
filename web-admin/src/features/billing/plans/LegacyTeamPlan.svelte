<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";

  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

<PlanContainer badge="Team (Legacy)" description="$250/mo flat + storage">
  {#snippet info()}
    $250/mo flat rate + storage overages. 10 GB included · $25/GB over. Up to 8
    slots.
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

  <p class="text-sm text-fg-tertiary mt-4 pb-4">
    Legacy flat-rate plan. $250/mo includes up to 8 slots and 10 GB storage,
    with $25/GB for overages. Upgrade to Pro for usage-based pricing.
  </p>
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
