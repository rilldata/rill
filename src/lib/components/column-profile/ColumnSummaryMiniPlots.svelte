<script lang="ts">
  import BarAndLabel from "$lib/components/BarAndLabel.svelte";
  import FormattedDataType from "$lib/components/data-types/FormattedDataType.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";

  import {
    formatInteger,
    formatCompactInteger,
    singleDigitPercentage,
  } from "$lib/util/formatters";
  import {
    CATEGORICALS,
    NUMERICS,
    TIMESTAMPS,
    DATA_TYPE_COLORS,
    BOOLEANS,
  } from "$lib/duckdb-data-types";

  import Histogram from "$lib/components/viz/histogram/SmallHistogram.svelte";
  import { TimestampSpark } from "../data-graphic/compositions/timestamp-profile";
  export let type;
  export let summary;
  export let totalRows: number;
  export let nullCount;
  export let example;
  export let view = "summaries"; // summaries, example
  export let containerWidth: number;

  // hide the null percentage number
  export let hideNullPercentage = false;
  export let compactBreakpoint = 350;

  $: exampleWidth =
    containerWidth > COLUMN_PROFILE_CONFIG.mediumCutoff
      ? COLUMN_PROFILE_CONFIG.exampleWidth.medium
      : COLUMN_PROFILE_CONFIG.exampleWidth.small;
  $: summaryWidthSize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[
      containerWidth < compactBreakpoint ? "small" : "medium"
    ];
  $: cardinalityFormatter =
    containerWidth > COLUMN_PROFILE_CONFIG.compactBreakpoint
      ? formatInteger
      : formatCompactInteger;

  /** used to convert a timestamp preview from the server for a sparkline. */
  function convertTimestampPreview(d) {
    return d.map((di) => {
      let pi = { ...di };
      pi.ts = new Date(pi.ts);
      return pi;
    });
  }

  const enum MiniPlotTypes {
    BAR_AND_LABEL,
    NUMERIC_HISTOGRAM,
    TIMESTAMP_SPARK, // a rollup sparkline
    TIMESTAMP_HISTOGRAM, // a legacy histogram timestamp histor
    NONE,
  }

  let miniPlotType: MiniPlotTypes = MiniPlotTypes.NONE;
  // check to see if the summary has cardinality. Otherwise do not show anything.
  if (totalRows > 0) {
    if (
      (CATEGORICALS.has(type) || BOOLEANS.has(type)) &&
      summary?.cardinality
    ) {
      miniPlotType = MiniPlotTypes.BAR_AND_LABEL;
    } else if (NUMERICS.has(type) && summary?.histogram?.length) {
      miniPlotType = MiniPlotTypes.NUMERIC_HISTOGRAM;
    } else if (TIMESTAMPS.has(type) && summary?.histogram?.length) {
      miniPlotType = MiniPlotTypes.TIMESTAMP_HISTOGRAM;
    } else if (TIMESTAMPS.has(type) && summary?.rollup?.spark?.length) {
      miniPlotType = MiniPlotTypes.TIMESTAMP_SPARK;
    }
  }
</script>

<div class="flex gap-2 items-center" class:hidden={view !== "summaries"}>
  <div class="flex items-center" style:width="{summaryWidthSize}px">
    {#if miniPlotType === MiniPlotTypes.BAR_AND_LABEL}
      <Tooltip location="right" alignment="center" distance={8}>
        <BarAndLabel
          color={DATA_TYPE_COLORS["VARCHAR"].bgClass}
          value={summary?.cardinality / totalRows}
        >
          |{cardinalityFormatter(summary?.cardinality)}|
        </BarAndLabel>
        <TooltipContent slot="tooltip-content">
          {formatInteger(summary?.cardinality)} unique values
        </TooltipContent>
      </Tooltip>
    {:else if miniPlotType === MiniPlotTypes.NUMERIC_HISTOGRAM}
      <Tooltip location="right" alignment="center" distance={8}>
        <Histogram
          data={summary.histogram}
          width={summaryWidthSize}
          height={18}
          fillColor={DATA_TYPE_COLORS["DOUBLE"].vizFillClass}
          baselineStrokeColor={DATA_TYPE_COLORS["DOUBLE"].vizStrokeClass}
        />
        <TooltipContent slot="tooltip-content">
          the distribution of the values of this column
        </TooltipContent>
      </Tooltip>
    {:else if miniPlotType === MiniPlotTypes.TIMESTAMP_SPARK}
      <Tooltip location="right" alignment="center" distance={8}>
        <TimestampSpark
          data={convertTimestampPreview(summary.rollup.spark)}
          xAccessor="ts"
          yAccessor="count"
          width={summaryWidthSize}
          height={18}
          top={0}
          bottom={0}
          left={0}
          right={0}
          leftBuffer={0}
          rightBuffer={0}
          area
          tweenIn
        />
        <TooltipContent slot="tooltip-content">the time series</TooltipContent>
      </Tooltip>
    {:else if miniPlotType === MiniPlotTypes.TIMESTAMP_HISTOGRAM}
      <Tooltip location="right" alignment="center" distance={8}>
        <Histogram
          data={summary.histogram}
          width={summaryWidthSize}
          height={18}
          fillColor={DATA_TYPE_COLORS["TIMESTAMP"].vizFillClass}
          baselineStrokeColor={DATA_TYPE_COLORS["TIMESTAMP"].vizStrokeClass}
        />

        <TooltipContent slot="tooltip-content">the time series</TooltipContent>
      </Tooltip>
    {/if}
  </div>

  <div
    style:width="{COLUMN_PROFILE_CONFIG.nullPercentageWidth}px"
    class:hidden={hideNullPercentage}
  >
    {#if totalRows !== 0 && totalRows !== undefined && nullCount !== undefined}
      <Tooltip location="right" alignment="center" distance={8}>
        <BarAndLabel
          showBackground={nullCount !== 0}
          color={DATA_TYPE_COLORS[type].bgClass}
          value={nullCount / totalRows || 0}
        >
          <span class:text-gray-300={nullCount === 0}
            >∅ {singleDigitPercentage(nullCount / totalRows)}</span
          >
        </BarAndLabel>
        <TooltipContent slot="tooltip-content">
          <svelte:fragment slot="title">
            what percentage of values are null?
          </svelte:fragment>
          {#if nullCount > 0}
            {singleDigitPercentage(nullCount / totalRows)} of the values are null
          {:else}
            no null values in this column
          {/if}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</div>
<Tooltip location="right" alignment="center" distance={8}>
  <div
    class:hidden={view !== "example"}
    class="
              pl-8 text-ellipsis overflow-hidden whitespace-nowrap text-right"
    style:max-width="{exampleWidth}px"
  >
    <FormattedDataType
      {type}
      isNull={example === null || example === ""}
      value={example}
    />
  </div>
  <TooltipContent slot="tooltip-content">
    <FormattedDataType
      value={example}
      {type}
      isNull={example === null || example === ""}
      dark
    />
  </TooltipContent>
</Tooltip>
