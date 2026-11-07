<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<a
  class="pricing-link"
  href="https://www.rilldata.com/pricing"
  target="_blank"
  rel="noreferrer noopener"
>
  {m.billing_see_pricing_details()}
  <svg
    class="w-3 h-3"
    viewBox="0 0 12 12"
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
  >
    <path d="M1 6h9M7.5 3l3 3-3 3" />
  </svg>
</a>

<style lang="postcss">
  .pricing-link {
    @apply mt-3 inline-flex items-center gap-1 self-end;
    @apply text-xs font-medium text-primary-600 no-underline;
  }

  .pricing-link:hover {
    @apply underline;
  }
</style>
