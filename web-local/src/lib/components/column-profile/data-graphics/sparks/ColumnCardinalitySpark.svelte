<script lang="ts">
  import {
    useRuntimeServiceGetCardinalityOfColumn,
    useRuntimeServiceGetTableCardinality,
  } from "@rilldata/web-common/runtime-client";
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
  export let compact = false;
  // export let compactBreakpoint = 350;

  $: cardinalityFormatter = !compact ? formatInteger : formatCompactInteger;

  let cardinality;

  $: cardinalityQuery = useRuntimeServiceGetCardinalityOfColumn(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: cardinality = $cardinalityQuery?.data?.categoricalSummary?.cardinality;

  /**
   * Get the total rows for this profile.
   */
  let totalRowsQuery;
  $: totalRowsQuery = useRuntimeServiceGetTableCardinality(
    $runtimeStore?.instanceId,
    objectName
  );
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
