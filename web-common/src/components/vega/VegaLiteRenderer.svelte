<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import type { ColorMapping } from "@rilldata/web-common/features/components/charts/types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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

  const runtimeClient = useRuntimeClient();

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let canvasDashboard = false;
  export let renderer: "canvas" | "svg" = "canvas";
  export let themeMode: "light" | "dark" = "light";
  export let config: Config | undefined = undefined;
  export let hasComparison: boolean = false;
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

  // Memoize colorMapping/hasComparison so they don't cause options to change
  // (which triggers svelte-vega to recreate the entire view, losing brush state).
  let stableColorMapping: ColorMapping = [];
  let stableHasComparison = false;
  $: if (JSON.stringify(colorMapping) !== JSON.stringify(stableColorMapping)) {
    stableColorMapping = colorMapping;
  }
  $: if (hasComparison !== stableHasComparison) {
    stableHasComparison = hasComparison;
  }

  $: options = createEmbedOptions({
    client: runtimeClient,
    width,
    height,
    config,
    renderer,
    themeMode,
    expressionFunctions,
    colorMapping: stableColorMapping,
    hasComparison: stableHasComparison,
  });

  // Create a more efficient key for component remounting
  $: configKey = config ? Object.keys(config).sort().join(",") : "default";
  $: colorMappingKey =
    stableColorMapping?.map((m) => `${m.value}:${m.color}`)?.join("|") ?? "";
  $: componentKey = `${themeMode}-${configKey}-${colorMappingKey}`;

  const onError = (e: Error) => {
    error = e.message;
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
  onmouseleave={handleMouseLeave}
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
        {onError}
      />
    {/key}
  {/if}
</div>
