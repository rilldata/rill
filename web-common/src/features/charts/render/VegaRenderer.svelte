<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { getRillTheme } from "@rilldata/web-common/features/charts/render/vega-config";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import { onMount } from "svelte";
  import {
    SignalListeners,
    Vega,
    View,
    VisualizationSpec,
    type EmbedOptions,
  } from "svelte-vega";
  import { ExpressionFunction } from "../types";
  //   import { VegaLiteTooltipHandler } from "./vega-tooltip";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let customDashboard = false;
  export let chartView = false;
  //   export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let viewVL: View;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let jwt = get(runtime).jwt;

  $: width = contentRect.width;
  $: height = contentRect.height * 0.95 - 80;

  //   $: if (viewVL && tooltipFormatter) {
  //     const handler = new VegaLiteTooltipHandler(tooltipFormatter);
  //     viewVL.tooltip(handler.handleTooltip);
  //     // https://stackoverflow.com/questions/59255654/vega-wont-update-until-the-mouse-has-brushed-over-the-div-containing-the-chart
  //     void viewVL.runAsync();
  //   }

  $: options = <EmbedOptions>{
    config: getRillTheme(),
    renderer: "svg",
    actions: false,
    logLevel: 0, // only show errors
    width: customDashboard ? width : undefined,
    expressionFunctions,
    height: chartView || !customDashboard ? undefined : height,
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

  function clearBrush() {
    if (viewVL) {
      viewVL.signal("brush", null);
      viewVL.data("brush_store", []);
      void viewVL.runAsync();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      clearBrush();
      dispatchEvent(new CustomEvent("brushCleared"));
    }
  }

  onMount(() => {
    window.addEventListener("keydown", handleKeydown);
    return () => {
      window.removeEventListener("keydown", handleKeydown);
    };
  });
</script>

<div
  bind:contentRect
  class:bg-white={customDashboard}
  class:px-4={customDashboard}
  class:pb-2={customDashboard}
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
