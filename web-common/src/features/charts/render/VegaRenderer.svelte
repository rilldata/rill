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
  import { afterUpdate, createEventDispatcher, onMount } from "svelte";
  import { getStateManagers } from "../../dashboards/state-managers/state-managers";
  import { PanDirection } from "../../dashboards/time-dimension-details/types";
  import PanLeftIcon from "@rilldata/web-common/components/icons/PanLeftIcon.svelte";
  import PanRightIcon from "@rilldata/web-common/components/icons/PanRightIcon.svelte";

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

  const dispatch = createEventDispatcher();
  const {
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
  } = getStateManagers();

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
    showControls = true;
  }

  function handleMouseLeave() {
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
  class="vega-renderer no-scrollbars size-full flex flex-col items-center justify-center relative mr-8"
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
      <div class="pan-controls absolute">
        {#if $canPanLeft}
          <button
            class="pan-button left w-8 h-8 pointer-events-auto"
            on:click={panLeft}
            aria-label="Pan left"
          >
            <PanLeftIcon />
          </button>
        {/if}
        {#if $canPanRight}
          <button
            class="pan-button right w-8 h-8 pointer-events-auto"
            on:click={panRight}
            aria-label="Pan right"
          >
            <PanRightIcon />
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

  .no-scrollbars {
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* Internet Explorer and Edge */
  }

  .no-scrollbars::-webkit-scrollbar {
    width: 0px;
    height: 0px;
    background: transparent; /* Chrome/Safari/Webkit */
  }

  .pan-controls {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
  }

  .pan-button {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: auto;
  }

  .pan-button.left {
    left: 10px;
  }

  .pan-button.right {
    right: -18px;
  }
</style>
