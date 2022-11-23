<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "../../application-config";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import { CATEGORICALS, NUMERICS, TIMESTAMPS } from "../../duckdb-data-types";
  import FormattedDataType from "../data-types/FormattedDataType.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  import ColumnCardinalitySpark from "./data-graphics/sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "./data-graphics/sparks/NullPercentageSpark.svelte";
  import NumericSpark from "./data-graphics/sparks/NumericSpark.svelte";

  import {
    useRuntimeServiceGetNullCount,
    useRuntimeServiceGetTableCardinality,
  } from "@rilldata/web-common/runtime-client";
  import TimestampSpark from "./data-graphics/sparks/TimestampSpark.svelte";

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

  /**
   * Get the null counts for this profile.
   */

  $: nullCountQuery = useRuntimeServiceGetNullCount(
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

  $: totalRowsQuery = useRuntimeServiceGetTableCardinality(
    $runtimeStore?.instanceId,
    objectName
  );
  let totalRows = 0;
  // FIXME: count should not be a string.
  $: totalRows = +$totalRowsQuery?.data?.cardinality;
</script>

<div class="flex gap-2 items-center" class:hidden={view !== "summaries"}>
  <div class="flex items-center" style:width="{summaryWidthSize}px">
    {#if totalRows}
      {#if CATEGORICALS.has(type)}
        <ColumnCardinalitySpark {objectName} {columnName} />
      {:else if NUMERICS.has(type)}
        <NumericSpark {objectName} {columnName} {containerWidth} />
      {:else if TIMESTAMPS.has(type) && false}
        <!-- no-op until the time series returns the right data -->
        <Tooltip location="right" alignment="center" distance={8}>
          <TimestampSpark
            {objectName}
            {columnName}
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
