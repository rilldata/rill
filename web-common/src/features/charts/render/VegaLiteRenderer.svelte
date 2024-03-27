<script lang="ts">
  import { getRillTheme } from "@rilldata/web-common/features/charts/render/vega-config";
  import { VegaLite, View, type EmbedOptions } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec; // VisualizationSpec;

  let options: EmbedOptions = {
    config: getRillTheme(),
    renderer: "svg",
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
