<script lang="ts">
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";

  export let leading: string | undefined = undefined;

  $: isEmbedded = EmbedStore.isEmbedded();
</script>

{#if !isEmbedded}
  <p class="text-sm text-fg-secondary text-center">
    {#if leading}{leading}{" "}{/if}Need help? Reach out to us on <a
      href="https://discord.gg/2ubRfjC7Rh"
      target="_blank">Discord</a
    >
  </p>
{/if}
