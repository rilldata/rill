<script lang="ts">
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import {
    formatCompactInteger,
    formatInteger,
  } from "@rilldata/web-local/lib/util/formatters";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let cardinality: number;
  export let totalRows: number;
  export let compact = false;

  $: cardinalityFormatter = !compact ? formatInteger : formatCompactInteger;
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
