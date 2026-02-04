<!-- @component
The NumericPlot component has three elements:
- toggles between the summary statistics and the top K values for the secondary plot
- a primary plot in the form of a histogram & rug plot
- a secondary plot in the form of a summary statistics or a top K plot
The goal is to make sure that even if the data isn't fetched, the component doesn't reflow once it does.
Otherwise, the page will jump around as the data is fetched.
-->
<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import { barplotPolyline } from "@rilldata/web-common/components/data-graphic/utils";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import SummaryStatistics from "@rilldata/web-common/components/icons/SummaryStatistics.svelte";
  import TopKIcon from "@rilldata/web-common/components/icons/TopK.svelte";
  import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    formatInteger,
    justEnoughPrecision,
  } from "@rilldata/web-common/lib/formatters";
  import type {
    NumericHistogramBinsBin,
    NumericOutliersOutlier,
    TopKEntry,
    V1NumericStatistics,
  } from "@rilldata/web-common/runtime-client";
  import { extent, bisector, max, min } from "d3-array";
  import { scaleLinear } from "d3-scale";
  import { interpolateBlues } from "d3-scale-chromatic";
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";
  import { fade, fly } from "svelte/transition";
  import SummaryNumberPlot from "./SummaryNumberPlot.svelte";
  import TopK from "./TopK.svelte";

  // Layout constants
  const histHeight = 64;
  const rugHeight = 16;
  const margins = { top: 1, right: 4, bottom: 0, left: 4 };

  export let data: NumericHistogramBinsBin[];
  export let rug: NumericOutliersOutlier[];
  export let summary: V1NumericStatistics | undefined;
  export let topK: TopKEntry[];
  export let totalRows: number;
  export let type: string;

  let summaryMode: "summary" | "topk" = "summary";
  let topKLimit = 15;
  let rowHeight = 24;
  let containerWidth = 400;
  let focusPoint: TopKEntry | undefined = undefined;

  $: if (summaryMode !== "summary") focusPoint = undefined;

  $: plotLeft = margins.left;
  $: plotRight = containerWidth - margins.right;
  $: plotTop = margins.top;
  $: plotBottom = histHeight - margins.bottom;

  // Histogram scales
  $: xMin = min(data, (d) => d.low);
  $: xMax = max(data, (d) => d.high);
  $: [, yMax] = extent(data, (d) => d.count);

  $: xScale = scaleLinear()
    .domain([xMin ?? 0, xMax ?? 1])
    .range([plotLeft, plotRight]);
  $: yScale = scaleLinear()
    .domain([0, yMax ?? 1])
    .range([plotBottom, plotTop]);

  // Histogram path
  $: separator = data?.length < 30 && INTEGERS.has(type) ? 1 : 0;
  const histGradientId = `hist-gradient-${guidGenerator()}`;
  $: histPath = data
    ? barplotPolyline(data, xScale, yScale, separator, false, 1)
    : "";

  // Bisection for hover
  const bisectLeft = bisector(
    (d: NumericHistogramBinsBin) => ((d.high ?? 0) + (d.low ?? 0)) / 2,
  );
  const bisect = (data: NumericHistogramBinsBin[], value: number) =>
    bisectLeft.left(data, value);

  // Mouse tracking
  let mouseX: number | undefined = undefined;
  $: hoveredBin =
    mouseX !== undefined && data
      ? data[bisect(data, xScale.invert(mouseX))]
      : undefined;

  // TopK focus tweened position
  const tweenedFocusX = tweened(0, { duration: 200, easing: cubicOut });
  const tweenedFocusY = tweened(0, { duration: 200, easing: cubicOut });
  $: if (focusPoint?.value) void tweenedFocusX.set(xScale(+focusPoint.value));
  $: if (focusPoint?.count !== undefined)
    void tweenedFocusY.set(yScale(focusPoint.count));

  // Rug rendering
  const rugTiers = 32;
  $: rugBuckets = buildRugBuckets(rug);

  function buildRugBuckets(
    rugData: NumericOutliersOutlier[],
  ): NumericOutliersOutlier[][] {
    if (!rugData?.length) return [];
    const largest = Math.max(...rugData.map((d) => d.count ?? 0));
    const buckets: NumericOutliersOutlier[][] = Array.from(
      { length: rugTiers },
      () => [],
    );
    rugData.forEach((datum) => {
      const count = datum.count ?? 0;
      if (count === 0) return;
      const idx = Math.min(
        rugTiers - 1,
        Math.floor((count / largest) * rugTiers),
      );
      buckets[idx] = [...buckets[idx], datum];
    });
    return buckets;
  }

  $: rugColorScale = scaleLinear().domain([0, 1]).range([0.2, 0.65]);

  function drawRugSegments(
    points: NumericOutliersOutlier[],
    scale: (v: number) => number,
  ): string {
    return points
      .map((p) => {
        const x = scale(p.high ?? 0);
        return `M${x},0 L${x},${rugHeight} L${x},0`;
      })
      .join("");
  }
</script>

<div
  role="group"
  on:mouseleave={() => {
    focusPoint = undefined;
  }}
>
  <div class="flex pb-4">
    <IconButton
      active={summaryMode === "summary"}
      tooltipLocation="top"
      marginClasses=""
      on:click={() => {
        summaryMode = "summary";
      }}
    >
      <SummaryStatistics />
      <svelte:fragment slot="tooltip-content">
        Show basic summary statistics
      </svelte:fragment>
    </IconButton>
    <IconButton
      tooltipLocation="top"
      active={summaryMode === "topk"}
      marginClasses=""
      on:click={() => {
        summaryMode = "topk";
      }}
    >
      <TopKIcon />
      <svelte:fragment slot="tooltip-content"
        >Show the top values</svelte:fragment
      >
    </IconButton>
  </div>
  <div bind:clientWidth={containerWidth}>
    <!-- Histogram -->

    <svg
      role="presentation"
      class="overflow-visible"
      width={containerWidth}
      height={histHeight}
      on:mousemove={(e) => {
        mouseX = e.offsetX;
      }}
      on:mouseleave={() => {
        mouseX = undefined;
      }}
    >
      <defs>
        <linearGradient id={histGradientId} x1="0" x2="0" y1="0" y2="1">
          <stop offset="5%" stop-color="var(--color-primary-600)" />
          <stop
            offset="95%"
            stop-color="var(--surface-background)"
            stop-opacity={0.4}
          />
        </linearGradient>
      </defs>

      {#if data}
        <!-- baseline -->
        <line
          x1={plotLeft}
          x2={plotRight}
          y1={plotBottom}
          y2={plotBottom}
          class="stroke-primary-400"
          stroke-width={1}
        />

        <!-- histogram bars -->
        {#if histPath?.length}
          <path d={histPath} fill="url(#{histGradientId})" />
          <path
            d={histPath}
            class="stroke-primary-400"
            fill="none"
            stroke-width={1}
          />
        {/if}

        <!-- zero line -->
        <line
          x1={xScale(0)}
          x2={xScale(0)}
          y1={yScale(0)}
          y2={plotTop}
          class="stroke-gray-400"
          stroke-dasharray="2,2"
          shape-rendering="crispEdges"
        />

        <!-- hover highlight -->
        {#if hoveredBin}
          <g
            transition:fade|global={{ duration: 50 }}
            shape-rendering="crispEdges"
          >
            <rect
              x={xScale(hoveredBin.low ?? 0)}
              y={plotTop}
              width={Math.abs(
                xScale(hoveredBin.high ?? 0) - xScale(hoveredBin.low ?? 0),
              )}
              height={plotBottom - plotTop}
              class="fill-gray-700"
              opacity={0.2}
            />
            <rect
              x={xScale(hoveredBin.low ?? 0)}
              y={yScale(hoveredBin.count ?? 0)}
              width={Math.abs(
                xScale(hoveredBin.high ?? 0) - xScale(hoveredBin.low ?? 0),
              )}
              height={plotBottom - yScale(hoveredBin.count ?? 0)}
              class="fill-primary-200"
            />
            <line
              x1={xScale(hoveredBin.low ?? 0)}
              x2={xScale(hoveredBin.low ?? 0)}
              y1={yScale(0)}
              y2={yScale(hoveredBin.count ?? 0)}
              class="stroke-primary-500"
              stroke-width="2"
            />
            <line
              x1={xScale(hoveredBin.high ?? 0)}
              x2={xScale(hoveredBin.high ?? 0)}
              y1={yScale(0)}
              y2={yScale(hoveredBin.count ?? 0)}
              class="stroke-primary-500"
              stroke-width="2"
            />
            <line
              x1={xScale(hoveredBin.low ?? 0)}
              x2={xScale(hoveredBin.high ?? 0)}
              y1={yScale(hoveredBin.count ?? 0)}
              y2={yScale(hoveredBin.count ?? 0)}
              class="stroke-primary-500"
              stroke-width="2"
            />
            <line
              x1={xScale(hoveredBin.low ?? 0)}
              x2={xScale(hoveredBin.high ?? 0)}
              y1={yScale(0) - 0.5}
              y2={yScale(0) - 0.5}
              class="stroke-primary-500"
              stroke-width={1}
            />
          </g>
        {/if}

        <!-- hover labels -->
        {#if hoveredBin?.low !== undefined}
          <g
            in:fly|global={{ duration: 200, x: -16 }}
            out:fly|global={{ duration: 200, x: -16 }}
            font-size="12"
            style:user-select={"none"}
          >
            <text
              x={plotLeft}
              y={plotTop + 12}
              class="fill-fg-secondary text-outline"
              opacity={0.8}
              >({justEnoughPrecision(hoveredBin?.low ?? 0)}, {justEnoughPrecision(
                hoveredBin?.high ?? 0,
              )}{hoveredBin?.high === data.at(-1)?.high ? ")" : "]"}</text
            >
            <text
              x={plotLeft}
              y={plotTop + 24}
              class="fill-fg-secondary text-outline"
              opacity={0.8}
            >
              {formatInteger(Math.trunc(hoveredBin.count ?? 0))} row{#if (hoveredBin.count ?? 0) !== 1}s{/if}
              ({(((hoveredBin.count ?? 0) / totalRows) * 100).toFixed(2)}%)
            </text>
          </g>
        {/if}

        <!-- topK focus indicator -->
        {#if focusPoint?.count !== undefined && focusPoint?.value && topK && summaryMode === "topk"}
          <g transition:fade={{ duration: 200 }}>
            <line
              x1={$tweenedFocusX}
              x2={$tweenedFocusX}
              y1={plotTop}
              y2={plotBottom}
              stroke="gray"
              stroke-width={1}
              opacity={0.7}
            />
            <line
              x1={$tweenedFocusX}
              x2={$tweenedFocusX}
              y1={$tweenedFocusY}
              y2={plotBottom}
              stroke="gray"
              stroke-width={6}
              opacity={0.7}
            />
          </g>
        {/if}
      {/if}
    </svg>

    <!-- Rug plot -->
    {#if rug}
      <svg class="overflow-visible" width={containerWidth} height={rugHeight}>
        <g transform="translate(0, 0)">
          {#each rugBuckets as bucket, i (i)}
            {#if bucket.length > 0}
              <path
                d={drawRugSegments(bucket, xScale)}
                stroke-width={1}
                stroke={interpolateBlues(rugColorScale(i / rugTiers))}
              />
            {/if}
          {/each}
        </g>
      </svg>
    {/if}

    <!-- Summary / TopK section -->
    <div
      class="pt-1"
      style:height={summaryMode === "summary" ? `${6 * rowHeight}px` : "auto"}
    >
      {#if summaryMode === "summary" && summary}
        {@const rowHeight = 24}
        <div class="pt-1">
          <SummaryNumberPlot
            {rowHeight}
            {summary}
            {type}
            {xScale}
            {plotRight}
          />
        </div>
      {:else if topK && summaryMode === "topk"}
        <div class="pt-1 px-1">
          <TopK
            onFocusTopK={(value) => {
              focusPoint = value;
            }}
            k={topKLimit}
            {topK}
            {totalRows}
            colorClass="bg-primary-200"
            {type}
          />
        </div>
      {/if}
    </div>
  </div>
</div>
