<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import { fade } from "svelte/transition";
  import { IconButton } from "../../../button";
  import { GraphicContext } from "../../../data-graphic/elements";
  import SimpleDataGraphic from "../../../data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithParentClientRect } from "../../../data-graphic/functional-components";
  import WithTween from "../../../data-graphic/functional-components/WithTween.svelte";
  import { HistogramPrimitive, Rug } from "../../../data-graphic/marks";
  import SummaryStatistics from "../../../icons/SummaryStatistics.svelte";
  import TopKIcon from "../../../icons/TopK.svelte";
  import SummaryNumberPlot from "./SummaryNumberPlot.svelte";
  import TopK from "./TopK.svelte";

  export let data;
  export let rug;
  export let summary;
  export let topK;
  export let totalRows;
  export let type;

  let summaryMode: "summary" | "topk" = "summary";

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
        show basic summary statistics
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
        >show the top values</svelte:fragment
      >
    </IconButton>
  </div>
  <WithParentClientRect let:rect let:styles let:toNumber>
    {#if data && summary}
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
        <SimpleDataGraphic let:config let:xScale let:yScale>
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
            separator={0}
          />

          <!-- support topK mouseover effect on graphs -->
          {#if focusPoint && summaryMode === "topk"}
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
        </SimpleDataGraphic>
        <SimpleDataGraphic top={0} bottom={0} height={16}>
          <Rug xAccessor="low" densityAccessor="count" data={rug} />
        </SimpleDataGraphic>
        {#if summaryMode === "summary"}
          <div class="pt-1">
            <SummaryNumberPlot
              min={summary?.min}
              max={summary?.max}
              mean={summary?.mean}
              q25={summary?.q25}
              q50={summary?.q50}
              q75={summary?.q75}
              {type}
            />
          </div>
        {:else if summaryMode === "topk"}
          <div class="pt-1 px-1">
            <TopK
              on:focus-top-k={(event) => {
                focusPoint = event.detail;
              }}
              {topK}
              {totalRows}
              colorClass="bg-red-200"
            />
          </div>
        {/if}
      </GraphicContext>
    {/if}
  </WithParentClientRect>
</div>
