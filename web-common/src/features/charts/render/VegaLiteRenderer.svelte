<script lang="ts">
  import { getVegaConfig } from "@rilldata/web-common/features/charts/render/vega-config";
  import { VegaLite, View } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec; // VisualizationSpec;

  // EmbedOptions type missing from svelte-vega
  interface Options {
    config: undefined | Record<string, unknown>;
    renderer: "canvas" | "svg";
  }

  let options: Options = {
    config: getVegaConfig(),
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
