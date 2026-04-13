<script lang="ts">
  import { page } from "$app/stores";

  let organization = $derived($page.params.organization);
</script>

<section class="usage-page">
  <h1 class="text-xl font-semibold text-fg-primary mb-2">Usage</h1>
  <p class="text-sm text-fg-secondary mb-6">
    View slot usage, storage consumption, and billing details for your
    organization.
  </p>

  <div class="coming-soon-card">
    <p class="text-fg-tertiary text-sm">
      Detailed usage metrics are coming soon.
    </p>
  </div>
</section>

<style lang="postcss">
  .usage-page {
    @apply max-w-4xl;
  }

  .coming-soon-card {
    @apply flex items-center justify-center border border-dashed rounded-xl p-12 bg-surface-background;
  }
</style>
