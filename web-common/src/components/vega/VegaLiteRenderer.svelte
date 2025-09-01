<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { onDestroy } from "svelte";
  import {
    type SignalListeners,
    VegaLite,
    type View,
    type VisualizationSpec,
  } from "svelte-vega";
  import type { Config } from "vega-lite";
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
  export let config: Config | undefined = undefined;
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let colorMapping: { value: string; color: string }[] = [];
  export let viewVL: View;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let tooltipHandler: VegaLiteTooltipHandler | null = null;

  $: width = contentRect.width;
  $: height = contentRect.height - 10;

  $: if (viewVL && tooltipFormatter) {
    // Clean up previous handler if it exists
    if (tooltipHandler) {
      tooltipHandler.destroy();
    }

    tooltipHandler = new VegaLiteTooltipHandler(tooltipFormatter);
    viewVL.tooltip(tooltipHandler.handleTooltip);
    void viewVL.runAsync();
  }

  $: options = createEmbedOptions({
    canvasDashboard,
    width,
    height,
    config,
    renderer,
    theme,
    expressionFunctions,
    colorMapping,
  });

  const onError = (e: CustomEvent<{ error: Error }>) => {
    error = e.detail.error.message;
  };

  const handleMouseLeave = () => {
    if (tooltipHandler) {
      tooltipHandler.removeTooltip();
    }
  };

  onDestroy(() => {
    if (tooltipHandler) {
      tooltipHandler.destroy();
      tooltipHandler = null;
    }
  });
</script>

<div
  bind:contentRect
  role="presentation"
  class:bg-surface={canvasDashboard}
  class:px-2={canvasDashboard}
  class="rill-vega-container overflow-y-auto overflow-x-hidden size-full flex flex-col items-center"
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
