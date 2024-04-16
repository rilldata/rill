<script lang="ts">
  import { getRillTheme } from "@rilldata/web-common/features/charts/render/vega-config";
  import {
    SignalListeners,
    VegaLite,
    View,
    type EmbedOptions,
  } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec; // VisualizationSpec;
  export let signalListeners: SignalListeners = {};

  let options: EmbedOptions = {
    config: getRillTheme(),
    renderer: "svg",
    actions: false,
    logLevel: 0, // only show errors
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
  <VegaLite
    {data}
    {spec}
    {signalListeners}
    {options}
    bind:view={viewVL}
    on:onError={onError}
  />
{/if}

<style lang="postcss">
  :global(.vega-embed) {
    width: 100%;
  }

  :global(#vg-tooltip-element) {
    @apply shadow-md;
    font-family: "Inter";
    font-size: 12px;
    background: rgba(250, 250, 250, 0.8);
  }

  :global(#vg-tooltip-element table tr td.value) {
    @apply font-normal;
  }
  :global(#vg-tooltip-element table tr td.key) {
    @apply text-gray-500;
  }
</style>
