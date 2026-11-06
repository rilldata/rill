<script lang="ts">
  import ArrowUpRight from "@rilldata/web-common/components/icons/ArrowUpRight.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
    {m.billing_view_detailed_usage()}
    <ArrowUpRight />
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
