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
</script>

{#if spec === undefined}
  <p>Chart not available</p>
{:else}
  <VegaEmbed
    {spec}
    {data}
    {signalListeners}
    options={vegaLiteOptions}
    bind:view
    on:onNewView
    on:onError
  />
{/if}
