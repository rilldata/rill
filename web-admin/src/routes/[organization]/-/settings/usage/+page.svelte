<script lang="ts">
  import { page } from "$app/stores";
  import type { PageData } from "./$types";
  import EmbeddedBillingPage from "../../../../../features/billing/usage/EmbeddedBillingPage.svelte";
  import OrgUsageTable from "@rilldata/web-admin/features/billing/usage/OrgUsageTable.svelte";

  let { data }: { data: PageData } = $props();

  let org = $derived($page.params.organization);
</script>

<section class="usage-page">
  <h1 class="text-xl font-semibold text-fg-primary mb-2">Usage</h1>

  <OrgUsageTable {org} />

  <EmbeddedBillingPage billingPortalUrl={data.billingPortalUrl} />
</section>

<style lang="postcss">
  .usage-page {
    @apply max-w-4xl;
  }
</style>
