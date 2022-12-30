<script lang="ts">
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { HistogramPrimitive } from "@rilldata/web-common/components/data-graphic/marks";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";

  export let compact = false;
  export let data;

  $: summaryWidthSize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];
</script>

{#if data}
  <Tooltip location="right" alignment="center" distance={8}>
    <SimpleDataGraphic
      xType="number"
      yType="number"
      yMin={0}
      width={summaryWidthSize}
      height={18}
      bodyBuffer={0}
      marginBuffer={0}
      left={1}
      right={1}
      top={4}
      bottom={1}
      let:config
    >
      <g class="text-red-200">
        <line
          x1={config.plotLeft}
          x2={config.plotRight}
          y1={config.plotBottom}
          y2={config.plotBottom}
          stroke="currentColor"
          stroke-width={0.5}
        />
      </g>
      <g class="text-red-300">
        <HistogramPrimitive
          {data}
          xLowAccessor="low"
          xHighAccessor="high"
          yAccessor="count"
          lineThickness={0.5}
          separator={0}
          color="currentColor"
          stopOpacity={0.5}
        />
      </g>
    </SimpleDataGraphic>
    <!-- <Histogram
      {data}
      width={summaryWidthSize}
      height={18}
      fillColor={DATA_TYPE_COLORS["DOUBLE"].vizFillClass}
      baselineStrokeColor={DATA_TYPE_COLORS["DOUBLE"].vizStrokeClass}
    /> -->
    <TooltipContent slot="tooltip-content">
      the distribution of the values of this column
    </TooltipContent>
  </Tooltip>
{/if}
