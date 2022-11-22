<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "../../application-config";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import {
    CATEGORICALS,
    DATA_TYPE_COLORS,
    NUMERICS,
    TIMESTAMPS,
  } from "../../duckdb-data-types";
  import { formatCompactInteger, formatInteger } from "../../util/formatters";
  import FormattedDataType from "../data-types/FormattedDataType.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  import CardinalitySpark from "./data-graphics/sparks/CardinalitySpark.svelte";

  import NullPercentageSpark from "./data-graphics/sparks/NullPercentageSpark.svelte";

  import {
    useRuntimeServiceGetNullCount,
    useRuntimeServiceTableCardinality,
  } from "@rilldata/web-common/runtime-client";
  import { convertTimestampPreview } from "../../util/convertTimestampPreview";
  import { TimestampSpark } from "../data-graphic/compositions/timestamp-profile";
  import Histogram from "../viz/histogram/SmallHistogram.svelte";
  import NumericSpark from "./data-graphics/sparks/NumericSpark.svelte";

  export let objectName: string;
  export let columnName: string;

  export let type;
  export let summary;
  // export let nullCount;
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

  /**
   * Get the null counts for this profile.
   */
  let nullCountQuery;
  $: if ($runtimeStore?.instanceId)
    nullCountQuery = useRuntimeServiceGetNullCount(
      $runtimeStore?.instanceId,
      objectName,
      columnName
    );

  let nullCount = 0;
  // FIXME: count should not be a string. For now, let's patch it.
  $: nullCount = +$nullCountQuery?.data?.count;

  /**
   * Get the total rows for this profile.
   */
  let totalRowsQuery;
  $: if ($runtimeStore?.instanceId) {
    totalRowsQuery = useRuntimeServiceTableCardinality(
      $runtimeStore?.instanceId,
      objectName
    );
  }
  let totalRows = 0;
  // FIXME: count should not be a string.
  $: totalRows = +$totalRowsQuery?.data?.cardinality;
</script>

<div class="flex gap-2 items-center" class:hidden={view !== "summaries"}>
  <div class="flex items-center" style:width="{summaryWidthSize}px">
    {#if totalRows}
      {#if CATEGORICALS.has(type)}
        <CardinalitySpark {objectName} {columnName} />
        <!-- <Tooltip location="right" alignment="center" distance={8}>
          <BarAndLabel
            color={DATA_TYPE_COLORS["VARCHAR"].bgClass}
            value={summary?.cardinality / totalRows}
          >
            |{cardinalityFormatter(summary?.cardinality)}|
          </BarAndLabel>
          <TooltipContent slot="tooltip-content">
            {formatInteger(summary?.cardinality)} unique values
          </TooltipContent>
        </Tooltip> -->
      {:else if NUMERICS.has(type)}
        <NumericSpark {objectName} {columnName} {containerWidth} />
        <!-- <Tooltip location="right" alignment="center" distance={8}>
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
        </Tooltip> -->
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
      <NullPercentageSpark {objectName} {columnName} />
      <!-- <Tooltip location="right" alignment="center" distance={8}>
        <BarAndLabel
          showBackground={nullCount !== 0}
          color={DATA_TYPE_COLORS[type]?.bgClass}
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
      </Tooltip> -->
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
