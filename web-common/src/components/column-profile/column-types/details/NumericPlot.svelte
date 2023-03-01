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
  import { outline } from "@rilldata/web-common/components/data-graphic/actions/outline";
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithParentClientRect } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithBisector from "@rilldata/web-common/components/data-graphic/functional-components/WithBisector.svelte";
  import WithTween from "@rilldata/web-common/components/data-graphic/functional-components/WithTween.svelte";
  import {
    HistogramPrimitive,
    Rug,
  } from "@rilldata/web-common/components/data-graphic/marks";
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
  import { cubicOut } from "svelte/easing";
  import { fade, fly } from "svelte/transition";
  import SummaryNumberPlot from "./SummaryNumberPlot.svelte";
  import TopK from "./TopK.svelte";

  export let data: NumericHistogramBinsBin[];
  export let rug: NumericOutliersOutlier[];
  export let summary: V1NumericStatistics;
  export let topK: TopKEntry[];
  export let totalRows: number;
  export let type: string;

  let summaryMode: "summary" | "topk" = "summary";

  let topKLimit = 15;

  // the rowHeight determines how big the secondary plots should be.
  // We will use this to predetermine the height of the secondary graphics;
  // this is important because we want to avoid reflowing the page after
  // the data has been fetched.
  let rowHeight = 24;

  let focusPoint = undefined;
  // reset focus point once the mode changes.
  $: if (summaryMode !== "summary") focusPoint = undefined;
</script>

<div
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
  <WithParentClientRect let:rect let:styles let:toNumber>
    <GraphicContext
      yMin={0}
      width={(rect?.width || 400) -
        toNumber(styles?.paddingLeft) -
        toNumber(styles?.paddingRight)}
      height={64}
      left={4}
      right={4}
      top={1}
      bottom={0}
      bodyBuffer={0}
      marginBuffer={0}
      xType="number"
      yType="number"
    >
      <SimpleDataGraphic let:config let:xScale let:yScale let:mouseoverValue>
        <WithBisector
          data={data || []}
          callback={(dt) => (dt.high + dt.low) / 2}
          value={mouseoverValue?.x}
          let:point
        >
          {#if data}
            <g class="text-red-400">
              <line
                x1={config.plotLeft}
                x2={config.plotRight}
                y1={config.plotBottom}
                y2={config.plotBottom}
                stroke="currentColor"
                stroke-width={1}
              />
            </g>
            <HistogramPrimitive
              {data}
              xLowAccessor="low"
              xHighAccessor="high"
              yAccessor="count"
              separator={data.length < 30 && INTEGERS.has(type) ? 1 : 0}
            />
            <!-- zero line -->
            <line
              x1={xScale(0)}
              x2={xScale(0)}
              y1={yScale(0)}
              y2={config.plotTop}
              class="stroke-gray-400"
              stroke-dasharray="2,2"
              shape-rendering="crispEdges"
            />
            <!-- show mouseover support shapes -->
            {#if point}
              <g
                transition:fade={{ duration: 50 }}
                shape-rendering="crispEdges"
              >
                <rect
                  x={xScale(point.low)}
                  y={config.plotTop}
                  width={Math.abs(xScale(point.high) - xScale(point.low))}
                  height={config.plotBottom - config.plotTop}
                  class="fill-gray-700"
                  opacity={0.2}
                />
                <rect
                  x={xScale(point.low)}
                  y={yScale(point.count)}
                  width={Math.abs(xScale(point.high) - xScale(point.low))}
                  height={config.plotBottom - yScale(point.count)}
                  class="fill-red-200"
                />
                <line
                  x1={xScale(point.low)}
                  x2={xScale(point.low)}
                  y1={yScale(0)}
                  y2={yScale(point.count)}
                  class="stroke-red-500"
                  stroke-width="2"
                />
                <line
                  x1={xScale(point.high)}
                  x2={xScale(point.high)}
                  y1={yScale(0)}
                  y2={yScale(point.count)}
                  class="stroke-red-500"
                  stroke-width="2"
                />
                <line
                  x1={xScale(point.low)}
                  x2={xScale(point.high)}
                  y1={yScale(point.count)}
                  y2={yScale(point.count)}
                  class="stroke-red-500"
                  stroke-width="2"
                />
                <line
                  x1={xScale(point.low)}
                  x2={xScale(point.high)}
                  y1={yScale(0) - 0.5}
                  y2={yScale(0) - 0.5}
                  class="stroke-red-500"
                  stroke-width={1}
                />
              </g>
            {/if}

            <!-- mouseovers -->
            {#if point?.low !== undefined}
              <g
                in:fly={{ duration: 200, x: -16 }}
                out:fly={{ duration: 200, x: -16 }}
                font-size={config.fontSize}
                style:user-select={"none"}
              >
                <text
                  use:outline
                  x={config.plotLeft}
                  y={config.plotTop + 12}
                  class="fill-gray-500"
                  opacity={0.8}
                  >({justEnoughPrecision(point?.low)}, {justEnoughPrecision(
                    point?.high
                  )}{point?.high === data.at(-1).high ? ")" : "]"}</text
                >
                <text
                  use:outline
                  x={config.plotLeft}
                  y={config.plotTop + 24}
                  class="fill-gray-500"
                  opacity={0.8}
                >
                  {formatInteger(~~point.count)} row{#if point.count !== 1}s{/if}
                  ({((point.count / totalRows) * 100).toFixed(2)}%)
                </text>
              </g>
            {/if}

            <!-- support topK mouseover effect on graphs -->
            {#if focusPoint && topK && summaryMode === "topk"}
              <g transition:fade|local={{ duration: 200 }}>
                <WithTween
                  value={[xScale(+focusPoint.value), yScale(focusPoint.count)]}
                  let:output
                  tweenProps={{ duration: 200, easing: cubicOut }}
                >
                  <line
                    x1={output[0]}
                    x2={output[0]}
                    y1={config.plotTop}
                    y2={config.plotBottom}
                    stroke="gray"
                    stroke-width={1}
                    opacity={0.7}
                  />
                  <line
                    x1={output[0]}
                    x2={output[0]}
                    y1={output[1]}
                    y2={config.plotBottom}
                    stroke="gray"
                    stroke-width={6}
                    opacity={0.7}
                  />
                </WithTween>
              </g>
            {/if}
          {/if}
        </WithBisector>
      </SimpleDataGraphic>
      <SimpleDataGraphic top={0} bottom={0} height={16}>
        {#if rug}
          <Rug xAccessor="high" densityAccessor="count" data={rug} />
        {/if}
      </SimpleDataGraphic>
      <!-- we'll prefill the height of the summary such that if the data hasn't been fetched yet,
        we'll preserve the space to keep the window from jumping.
      -->
      <div
        class="pt-1"
        style:height={summaryMode === "summary" ? `${6 * rowHeight}px` : "auto"}
      >
        {#if summaryMode === "summary" && summary}
          {@const rowHeight = 24}
          <div class="pt-1">
            <SummaryNumberPlot
              {rowHeight}
              min={summary?.min}
              max={summary?.max}
              mean={summary?.mean}
              q25={summary?.q25}
              q50={summary?.q50}
              q75={summary?.q75}
              {type}
            />
          </div>
        {:else if topK && summaryMode === "topk"}
          <div class="pt-1 px-1">
            <TopK
              on:focus-top-k={(event) => {
                focusPoint = event.detail;
              }}
              k={topKLimit}
              {topK}
              {totalRows}
              colorClass="bg-red-200"
              {type}
            />
          </div>
        {/if}
      </div>
    </GraphicContext>
  </WithParentClientRect>
</div>
