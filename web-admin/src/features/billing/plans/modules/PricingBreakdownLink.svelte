<script lang="ts">
  import ArrowUpRight from "@rilldata/web-common/components/icons/ArrowUpRight.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<a
  class="pricing-breakdown-link"
  href="https://www.rilldata.com/pricing#full-comparison"
  target="_blank"
  rel="noreferrer noopener"
>
  {m.billing_view_pricing_breakdown()}
  <ArrowUpRight />
</a>

<style lang="postcss">
  .pricing-breakdown-link {
    @apply inline-flex items-center gap-1;
    @apply text-sm font-medium text-primary-600 no-underline;
  }

  .pricing-breakdown-link:hover {
    @apply underline;
  }
</style>
