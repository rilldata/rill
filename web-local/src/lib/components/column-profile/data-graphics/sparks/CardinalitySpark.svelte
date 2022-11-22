<script lang="ts">
  import {
    runtimeServiceGetCardinalityOfColumn,
    useRuntimeServiceTableCardinality,
  } from "@rilldata/web-common/runtime-client";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import {
    formatCompactInteger,
    formatInteger,
  } from "@rilldata/web-local/lib/util/formatters";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let objectName: string;
  export let columnName: string;
  export let containerWidth: number = 500;
  // export let compactBreakpoint = 350;

  // $: exampleWidth =
  //   containerWidth > COLUMN_PROFILE_CONFIG.mediumCutoff
  //     ? COLUMN_PROFILE_CONFIG.exampleWidth.medium
  //     : COLUMN_PROFILE_CONFIG.exampleWidth.small;
  // $: summaryWidthSize =
  //   COLUMN_PROFILE_CONFIG.summaryVizWidth[
  //     containerWidth < compactBreakpoint ? "small" : "medium"
  //   ];
  $: cardinalityFormatter =
    containerWidth > COLUMN_PROFILE_CONFIG.compactBreakpoint
      ? formatInteger
      : formatCompactInteger;

  // FIXME: runtimeServiceGetCardinalityOfColumn seems to return a promise.
  let cardinalityQuery;
  let cardinality;

  $: if ($runtimeStore?.instanceId)
    cardinalityQuery = runtimeServiceGetCardinalityOfColumn(
      $runtimeStore?.instanceId,
      objectName,
      columnName
    ).then((result) => {
      cardinality = +result.cardinality;
    });
  //$: if (cardinalityQuery) cardinality = +$cardinalityQuery?.data?.count;

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
  let totalRows;
  // FIXME: count should not be a string.
  $: totalRows = +$totalRowsQuery?.data?.cardinality;
</script>

{#if cardinality && totalRows}
  <Tooltip location="right" alignment="center" distance={8}>
    <BarAndLabel
      color={DATA_TYPE_COLORS["VARCHAR"].bgClass}
      value={totalRows > 0 && totalRows !== undefined
        ? cardinality / totalRows
        : 0}
    >
      <span>
        |{cardinalityFormatter(cardinality)}|
      </span>
    </BarAndLabel>
    <TooltipContent slot="tooltip-content">
      {formatInteger(cardinality)} unique values
    </TooltipContent>
  </Tooltip>
{/if}
