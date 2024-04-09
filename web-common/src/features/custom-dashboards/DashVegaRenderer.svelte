<script lang="ts">
  import { VegaLite, View, VisualizationSpec } from "svelte-vega";
  import { getRillTheme } from "../charts/render/vega-config";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let height: number;
  export let width: number;

  let viewVL: View;
  let error = "";

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };
</script>

{#if error}
  <p>{error}</p>
{:else}
  <VegaLite
    {data}
    {spec}
    bind:view={viewVL}
    on:onError={onError}
    options={{
      actions: false,
      height,
      width,
      config: getRillTheme(),
    }}
  />
{/if}

<style>
  :global(.vega-embed) {
    width: 100% !important;
  }
</style>
