<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { getRillTheme } from "@rilldata/web-common/features/charts/render/vega-config";
  import {
    SignalListeners,
    VegaLite,
    View,
    type EmbedOptions,
    VisualizationSpec,
  } from "svelte-vega";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let error: string | null = null;
  export let customDashboard = false;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let viewVL: View;

  $: width = contentRect.width;
  $: height = contentRect.height * 0.8;

  $: options = <EmbedOptions>{
    config: getRillTheme(),
    renderer: "svg",
    actions: false,
    logLevel: 0, // only show errors
    width: customDashboard ? width : undefined,
    height: customDashboard ? height : undefined,
  };

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };
</script>

<div
  bind:contentRect
  class:bg-white={customDashboard}
  class:p-4={customDashboard}
  class="overflow-hidden size-full flex flex-col items-center justify-center"
>
  {#if error}
    <div
      class="size-full text-[3.2em] flex flex-col items-center justify-center gap-y-2"
    >
      <CancelCircle />
      {error}
    </div>
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
</div>

<style>
  :global(.vega-embed) {
    width: 100%;
  }
</style>
