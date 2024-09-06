<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { getRillTheme } from "@rilldata/web-common/features/charts/render/vega-config";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import {
    SignalListeners,
    Vega,
    View,
    VisualizationSpec,
    type EmbedOptions,
  } from "svelte-vega";
  import { ExpressionFunction, VLTooltipFormatter } from "../types";
  import { VegaLiteTooltipHandler } from "./vega-tooltip";
  import { createEventDispatcher } from "svelte";
  import { getStateManagers } from "../../dashboards/state-managers/state-managers";

  export let data: Record<string, unknown> = {};
  export let spec: VisualizationSpec;
  export let signalListeners: SignalListeners = {};
  export let expressionFunctions: ExpressionFunction = {};
  export let error: string | null = null;
  export let customDashboard = false;
  export let chartView = false;
  export let tooltipFormatter: VLTooltipFormatter | undefined = undefined;
  export let view: View;

  let contentRect = new DOMRect(0, 0, 0, 0);
  let jwt = get(runtime).jwt;

  type PanDirection = "left" | "right";

  const dispatch = createEventDispatcher();
  const {
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
  } = getStateManagers();

  $: console.log("tdd alternate chart", {
    canPanLeft: $canPanLeft,
    canPanRight: $canPanRight,
    getNewPanRange: $getNewPanRange,
  });

  $: width = contentRect.width;
  $: height = contentRect.height * 0.95 - 80;

  $: if (view && tooltipFormatter) {
    const handler = new VegaLiteTooltipHandler(tooltipFormatter);
    view.tooltip(handler.handleTooltip);
    // https://stackoverflow.com/questions/59255654/vega-wont-update-until-the-mouse-has-brushed-over-the-div-containing-the-chart
    void view.runAsync();
  }

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

  function panCharts(direction: PanDirection) {
    const panRange = $getNewPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;
    dispatch("pan", { start, end });
  }

  let showControls = false;

  function handleMouseEnter() {
    console.log("handleMouseEnter");
    showControls = true;
  }

  function handleMouseLeave() {
    console.log("handleMouseLeave");
    showControls = false;
  }

  function panLeft() {
    panCharts("left");
  }

  function panRight() {
    panCharts("right");
  }
</script>

<div
  bind:contentRect
  class:bg-white={customDashboard}
  class:px-4={customDashboard}
  class:pb-2={customDashboard}
  class="overflow-hidden size-full flex flex-col items-center justify-center"
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  role="figure"
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
    {#if showControls}
      <div class="pan-controls">
        {#if $canPanLeft}
          <button class="pan-button left" on:click={panLeft}>
            <svg viewBox="0 0 33 33" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M9.335 16.795L21.678 5.756C22.129 5.352 22.844 5.672 22.844 6.277L22.844 27.342C22.844 27.948 22.128 28.268 21.677 27.863L9.335 16.795Z"
              />
            </svg>
          </button>
        {/if}
        {#if $canPanRight}
          <button class="pan-button right" on:click={panRight}>
            <svg viewBox="0 0 33 33" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M24.265 16.805L11.922 27.844C11.471 28.248 10.756 27.928 10.756 27.323L10.756 6.258C10.756 5.652 11.472 5.332 11.923 5.737L24.265 16.805Z"
              />
            </svg>
          </button>
        {/if}
      </div>
    {/if}
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
