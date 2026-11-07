<script lang="ts">
  import type { PageData } from "./$types";
  import EmbeddedBillingPage from "@rilldata/web-admin/features/billing/EmbeddedBillingPage.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { data }: { data: PageData } = $props();
</script>

<section class="usage-page">
  <h1 class="text-xl font-semibold text-fg-primary mb-2">
    {m.billing_usage_title()}
  </h1>

  <EmbeddedBillingPage billingPortalUrl={data.billingPortalUrl} />
</section>

<style lang="postcss">
  .usage-page {
    @apply max-w-4xl;
  }
</style>
