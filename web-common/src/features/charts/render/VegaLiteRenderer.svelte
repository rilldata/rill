<script lang="ts">
  import { VegaLite, View } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec; // VisualizationSpec;

  let viewVL: View;
  $: error = "";

  const onError = (e) => {
    e = e.detail.error.message;
  };
</script>

{#if error}
  <p>{error}</p>
{:else}
  <VegaLite {data} {spec} bind:view={viewVL} on:onError={onError} />
{/if}

<style>
  :global(.vega-embed) {
    width: 100%;
  }
</style>
