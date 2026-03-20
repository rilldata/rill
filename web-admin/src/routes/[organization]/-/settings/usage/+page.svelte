<script lang="ts">
  import RillEmbed from "@rilldata/web-admin/features/embeds/RillEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
</script>

<section class="usage-section">
  <h3 class="section-title">Usage</h3>
  <RillEmbed
    org={data.organization}
    resource="embedded_usage"
    height="600px"
  />
</section>

<style lang="postcss">
  .usage-section {
    @apply border border-border rounded-lg p-5 mb-6;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide mb-3;
  }
</style>
