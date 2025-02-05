<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy } from "svelte";
  import {
    type EmbedOptions,
    type SignalListeners,
    Vega,
    type View,
    type VisualizationSpec,
  } from "svelte-vega";
  import { get } from "svelte/store";
  import type { ExpressionFunction, VLTooltipFormatter } from "./types";
  import { getRillTheme } from "./vega-config";
  import { VegaLiteTooltipHandler } from "./vega-tooltip";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let canvasDashboard = false;
  export let chartView = false;
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let view: View;
  export let isScrubbing: boolean;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let jwt = get(runtime).jwt;

  $: width = contentRect.width;
  $: height = contentRect.height * 0.95 - 80;

  let tooltipTimer: number | null = null;
  const TOOLTIP_DELAY = 200;

  function createHoverIntentTooltipHandler(baseHandler: any) {
    return function (handler: any, event: MouseEvent, item: any, value: any) {
      if (!event || isScrubbing) {
        return;
      }
      if (event.type === "pointermove") {
        if (tooltipTimer !== null) {
          clearTimeout(tooltipTimer);
        }
        tooltipTimer = window.setTimeout(() => {
          baseHandler.call(this, handler, event, item, value);
        }, TOOLTIP_DELAY);
      } else if (event.type === "pointerout") {
        if (tooltipTimer !== null) {
          clearTimeout(tooltipTimer);
          tooltipTimer = null;
        }
        baseHandler.call(this, handler, event, item, null);
      }
    };
  }

  $: if (view && tooltipFormatter) {
    const handler = new VegaLiteTooltipHandler(tooltipFormatter);
    view.tooltip(createHoverIntentTooltipHandler(handler.handleTooltip));
    void view.runAsync();
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

  onDestroy(() => {
    if (tooltipTimer !== null) {
      clearTimeout(tooltipTimer);
    }
  });
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
    <Vega
      {data}
      {spec}
      {signalListeners}
      {options}
      bind:view
      on:onError={onError}
    />
  {/if}
</div>

<style lang="postcss">
  :global(.vega-embed) {
    width: 100%;
  }

  :global(#vg-tooltip-element, #rill-vg-tooltip) {
    @apply absolute border border-slate-300 p-3 rounded-lg pointer-events-none;
    background: rgba(255, 255, 255, 0.9);
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
