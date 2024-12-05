<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { getRillTheme } from "@rilldata/web-common/features/canvas-components/render/vega-config";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    type EmbedOptions,
    type SignalListeners,
    VegaLite,
    type View,
    type VisualizationSpec,
  } from "svelte-vega";
  import { get } from "svelte/store";
  import type { ExpressionFunction, VLTooltipFormatter } from "../types";
  import { VegaLiteTooltipHandler } from "./vega-tooltip";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let canvasDashboard = false;
  export let chartView = false;
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let viewVL: View;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let jwt = get(runtime).jwt;

  $: width = contentRect.width;
  $: height = contentRect.height * 0.95 - 80;

  $: if (viewVL && tooltipFormatter) {
    const handler = new VegaLiteTooltipHandler(tooltipFormatter);
    viewVL.tooltip(handler.handleTooltip);
    void viewVL.runAsync();
  }

  $: options = <EmbedOptions>{
    config: getRillTheme(canvasDashboard),
    renderer: "svg",
    actions: false,
    logLevel: 0, // only show errors
    width: canvasDashboard ? width : undefined,
    expressionFunctions,
    height: chartView || !canvasDashboard ? undefined : height,
    loader: {
      baseURL: `${get(runtime).host}/v1/instances/${get(runtime).instanceId}/assets/`,
      ...(jwt &&
        jwt.token && {
          http: {
            headers: {
              Authorization: `Bearer ${jwt.token}`,
            },
          },
        }),
    },
  };

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };
</script>

<div
  bind:contentRect
  class:bg-white={canvasDashboard}
  class:px-4={canvasDashboard}
  class:pb-2={canvasDashboard}
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

<style lang="postcss">
  :global(.vega-embed) {
    width: 100%;
  }

  :global(#rill-vg-tooltip) {
    @apply absolute border border-slate-300 p-3 rounded-lg pointer-events-none;
    background: white;
    & h2 {
      @apply text-slate-500 text-sm font-semibold mb-2;
    }

    & table {
      @apply border-spacing-0;
    }

    & td {
      @apply truncate py-0.5;
    }

    & td.key {
      @apply text-left px-1 font-normal truncate;
      max-width: 250px;
    }

    & td.value {
      @apply text-left truncate font-semibold ui-copy-number;
      max-width: 250px;
    }
  }
</style>
