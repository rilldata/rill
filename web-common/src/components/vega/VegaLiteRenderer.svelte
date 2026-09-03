<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import type { ColorMapping } from "@rilldata/web-common/features/components/charts/types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy } from "svelte";
  import {
    type EmbedOptions,
    type SignalListeners,
    VegaLite,
    type View,
    type VisualizationSpec,
  } from "svelte-vega";
  import type { Config } from "vega-lite";
  import type { ExpressionFunction, VLTooltipFormatter } from "./types";
  import { createBaseEmbedOptions } from "./vega-embed-options";
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
  // The container is measured asynchronously after mount. The view is only embedded once a
  // real size is known, so the first render is already at the right dimensions.
  let measured = false;

  $: width = contentRect.width;
  $: height = contentRect.height - 10;
  $: if (width > 0) measured = true;

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

  // Deliberately excludes width/height so a resize leaves this object's identity untouched
  // and svelte-vega resizes the existing view rather than re-embedding it.
  $: baseOptions = createBaseEmbedOptions({
    client: runtimeClient,
    config,
    renderer,
    themeMode,
    expressionFunctions,
    colorMapping: stableColorMapping,
    hasComparison: stableHasComparison,
  });

  // svelte-vega tears down and re-embeds the view (a full Vega-Lite compile, parse and
  // render) whenever `options` changes: it keeps the previous options in a Svelte `$state`
  // proxy, so its `===` comparison of nested objects such as `config` and `tooltip` never
  // passes and its cheap `view.width()` path is unreachable. Sizing the view through `options`
  // therefore re-embedded every chart on every frame of a divider drag, which stalled the page
  // and flashed blank canvases. Instead, `options` only picks up the size when it is (re)built
  // and later size changes are applied to the live view below.
  // `withEmbedSize` reads `width` and `height` inside a function on purpose: a `$:` statement
  // only tracks the variables it references directly.
  $: options = withEmbedSize(baseOptions, measured);

  function withEmbedSize(base: EmbedOptions, hasSize: boolean): EmbedOptions {
    if (!hasSize) return base;
    return { ...base, width, height };
  }

  // Resize the existing view in place. vega-embed applies `options.width`/`height` the same way,
  // so this matches the initial embed. Setting an unchanged size is a no-op for Vega.
  $: if (viewVL && measured) {
    viewVL.width(width);
    viewVL.height(height);
    void viewVL.runAsync();
  }

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
      {#if measured}
        <VegaLite
          {data}
          {spec}
          {signalListeners}
          {options}
          bind:view={viewVL}
          {onError}
        />
      {/if}
    {/key}
  {/if}
</div>
