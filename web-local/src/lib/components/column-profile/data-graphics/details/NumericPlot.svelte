<script lang="ts">
  import { GraphicContext } from "../../../data-graphic/elements";
  import SimpleDataGraphic from "../../../data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithParentClientRect } from "../../../data-graphic/functional-components";
  import { Rug } from "../../../data-graphic/guides";
  import HistogramPrimitive from "../../../data-graphic/marks/HistogramPrimitive.svelte";
  import SummaryNumberPlot from "./SummaryNumberPlot.svelte";

  export let data;
  export let rug;
  export let summary;
  export let plotPad = 24;
</script>

<WithParentClientRect let:rect>
  {#if data && summary}
    <GraphicContext
      yMin={0}
      width={(rect?.width || 400) - 16}
      height={64}
      left={8}
      right={8}
      top={0}
      bottom={0}
      bodyBuffer={0}
      marginBuffer={0}
      xType="number"
      yType="number"
    >
      <SimpleDataGraphic let:config>
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
      </SimpleDataGraphic>
      <!-- <NumericHistogram
      width={(rect?.width || 400) - plotPad}
      height={65}
      {data}
      min={summary?.min}
      qlow={summary?.q25}
      median={summary?.q50}
      qhigh={summary?.q75}
      mean={summary?.mean}
      max={summary?.max}
    /> -->
      <SimpleDataGraphic>
        <Rug data={rug} />
      </SimpleDataGraphic>
      <SummaryNumberPlot
        min={summary?.min}
        max={summary?.max}
        mean={summary?.mean}
        q25={summary?.q25}
        q50={summary?.q50}
        q75={summary?.q75}
      />
    </GraphicContext>
  {/if}
  <!-- {#if rug && rug?.length}
    <OutlierHistogram
      width={(rect?.width || 400) - plotPad}
      height={15}
      data={rug}
      mean={summary?.mean}
      sd={summary?.sd}
      min={summary?.min}
      max={summary?.max}
    />
  {/if} -->
</WithParentClientRect>
