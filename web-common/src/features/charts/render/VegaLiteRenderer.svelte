<script lang="ts">
  import { VegaLite, View } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec; // VisualizationSpec;

  // EmbedOptions type missing from svelte-vega
  interface Options {
    theme: undefined | "vox" | "ggplot2";
  }

  let options: Options = {
    theme: "vox",
  };

  let viewVL: View;
  $: error = "";

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };
</script>

{#if error}
  <p>{error}</p>
{:else}
  <VegaLite {data} {spec} {options} bind:view={viewVL} on:onError={onError} />
{/if}

<style>
  :global(.vega-embed) {
    width: 100%;
  }
</style>
