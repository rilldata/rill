<script lang="ts">
  import ConnectClientButtons from "./ConnectClientButtons.svelte";
</script>

<div class="connect-hero">
  <div class="connect-hero-title">
    Query your Rill metrics from your favorite AI client.
  </div>
  <div class="connect-hero-subtitle">Or directly ask questions in Rill.</div>
  <ConnectClientButtons />
</div>

<style lang="postcss">
  .connect-hero {
    @apply mt-8 w-full max-w-md;
    @apply flex flex-col items-center gap-2;
    @apply mx-auto p-5;
    @apply rounded-lg border border-border bg-surface-base;
  }

  .connect-hero-title {
    @apply text-sm font-semibold text-fg-primary text-center;
  }

  .connect-hero-subtitle {
    @apply mb-2 text-xs text-fg-secondary text-center;
  }
</style>
