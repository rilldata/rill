<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { onDestroy } from "svelte";
  import {
    type SignalListeners,
    Vega,
    type View,
    type VisualizationSpec,
  } from "svelte-vega";
  import type { ExpressionFunction, VLTooltipFormatter } from "./types";
  import { createEmbedOptions } from "./vega-embed-options";
  import { VegaLiteTooltipHandler } from "./vega-tooltip";
  import "./vega.css";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let canvasDashboard = false;
  export let renderer: "canvas" | "svg" = "canvas";
  export let theme: "light" | "dark" = "light";
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let view: View;
  export let isScrubbing: boolean;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let tooltipHandler: VegaLiteTooltipHandler | null = null;

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
    if (tooltipHandler) {
      tooltipHandler.destroy();
    }

    tooltipHandler = new VegaLiteTooltipHandler(tooltipFormatter);
    view.tooltip(createHoverIntentTooltipHandler(tooltipHandler.handleTooltip));
    void view.runAsync();
  }

  $: options = createEmbedOptions({
    canvasDashboard,
    width,
    height,
    renderer,
    themeMode: theme,
    colorMapping: [],
    expressionFunctions,
    useExpressionInterpreter: false,
  });

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };

  const handleMouseLeave = () => {
    if (tooltipTimer !== null) {
      clearTimeout(tooltipTimer);
      tooltipTimer = null;
    }
    if (tooltipHandler) {
      tooltipHandler.removeTooltip();
    }
  };

  onDestroy(() => {
    if (tooltipTimer !== null) {
      clearTimeout(tooltipTimer);
    }
    if (tooltipHandler) {
      tooltipHandler.destroy();
      tooltipHandler = null;
    }
  });
</script>

<div
  bind:contentRect
  role="presentation"
  class:px-4={canvasDashboard}
  class:pb-2={canvasDashboard}
  class="rill-vega-container overflow-hidden size-full flex flex-col items-center justify-center"
  on:mouseleave={handleMouseLeave}
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
