<script lang="ts">
  import type { EmbedOptions, Mode } from "vega-embed";
  import VegaEmbed from "./VegaEmbed.svelte";
  import type { SignalListeners, View } from "./types";

  export let spec; //VisualizationSpec
  export let options: EmbedOptions = {};
  export let data: Record<string, unknown> = {};
  export let signalListeners: SignalListeners = {};
  export let view: View | undefined = undefined;

  const mode = "vega-lite" as Mode;
  $: vegaLiteOptions = { ...options, mode: mode };

  let error = "";
</script>

{#if error}
  <p>{error}</p>
{:else}
  <VegaEmbed
    {spec}
    {data}
    {signalListeners}
    options={vegaLiteOptions}
    bind:view
    on:onNewView
    on:onError={(e) => (error = e.detail)}
  />
{/if}
