<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import type { ColorMapping } from "@rilldata/web-common/features/components/charts/types";
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
  export let themeMode: "light" | "dark" = "light";
  export let config: Config | undefined = undefined;
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let colorMapping: ColorMapping = [];
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
    themeMode,
    expressionFunctions,
    colorMapping,
  });

  // Create a more efficient key for component remounting
  $: configKey = config ? Object.keys(config).sort().join(",") : "default";
  $: colorMappingKey =
    colorMapping?.map((m) => `${m.value}:${m.color}`).join("|") ?? "";
  $: componentKey = `${themeMode}-${configKey}-${colorMappingKey}`;

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
    {#key componentKey}
      <VegaLite
        {data}
        {spec}
        {signalListeners}
        {options}
        bind:view={viewVL}
        on:onError={onError}
      />
    {/key}
  {/if}
</div>
