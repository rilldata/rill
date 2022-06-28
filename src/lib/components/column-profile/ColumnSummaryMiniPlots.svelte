<script lang="ts">
  import BarAndLabel from "$lib/components/BarAndLabel.svelte";
  import FormattedDataType from "$lib/components/data-types/FormattedDataType.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";

  import {
    formatCompactInteger,
    formatInteger,
    singleDigitPercentage,
  } from "$lib/util/formatters";
  import {
    BOOLEANS,
    CATEGORICALS,
    DATA_TYPE_COLORS,
    NUMERICS,
    TIMESTAMPS,
  } from "$lib/duckdb-data-types";

  import Histogram from "$lib/components/viz/histogram/SmallHistogram.svelte";
  import { TimestampSpark } from "../data-graphic/compositions/timestamp-profile";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview.js";

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
</script>

<div class="flex gap-2 items-center" class:hidden={view !== "summaries"}>
  <div class="flex items-center" style:width="{summaryWidthSize}px">
    {#if totalRows}
      {#if (CATEGORICALS.has(type) || BOOLEANS.has(type)) && summary?.cardinality}
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
      {:else if NUMERICS.has(type) && summary?.histogram?.length}
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
      {:else if TIMESTAMPS.has(type) /** a legacy histogram type or a new rollup spark */ && (summary?.histogram?.length || summary?.rollup?.spark?.length)}
        <Tooltip location="right" alignment="center" distance={8}>
          {#if summary?.rollup?.spark}
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
          {:else}
            <Histogram
              data={summary.histogram}
              width={summaryWidthSize}
              height={18}
              fillColor={DATA_TYPE_COLORS["TIMESTAMP"].vizFillClass}
              baselineStrokeColor={DATA_TYPE_COLORS["TIMESTAMP"].vizStrokeClass}
            />
          {/if}
          <TooltipContent slot="tooltip-content">
            the time series
          </TooltipContent>
        </Tooltip>
      {/if}
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
