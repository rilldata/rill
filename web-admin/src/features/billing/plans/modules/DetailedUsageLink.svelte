<script lang="ts">
  let {
    billingPortalUrl,
  }: {
    billingPortalUrl: string | undefined;
  } = $props();
</script>

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

<style lang="postcss">
  .pricing-link-top {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-link-top:hover {
    @apply underline;
  }
</style>
