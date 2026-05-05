<script lang="ts">
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";

  $: isEmbedded = EmbedStore.isEmbedded();
</script>

{#if !isEmbedded}
  <p class="text-sm text-fg-secondary text-center">
    Need help? Reach out to us on <a
      href="https://discord.gg/2ubRfjC7Rh"
      target="_blank">Discord</a
    >
  </p>
{/if}
