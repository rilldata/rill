<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-common/layout/config";
  import {
    DATA_TYPE_COLORS,
    isNested,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    formatCompactInteger,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let cardinality: number;
  export let totalRows: number;
  export let compact = false;
  export let type = "VARCHAR";

  $: cardinalityFormatter = !compact
    ? formatCompactInteger
    : formatCompactInteger;

  $: innerType = isNested(type) ? "STRUCT" : type;
</script>

{#if cardinality && totalRows}
  <Tooltip location="right" alignment="center" distance={8}>
    <BarAndLabel
      compact
      color={DATA_TYPE_COLORS[innerType].bgClass}
      value={totalRows > 0 && totalRows !== undefined
        ? cardinality / totalRows
        : 0}
    >
      <span
        style:font-size="{COLUMN_PROFILE_CONFIG.fontSize}px"
        class="ui-copy-number"
      >
        {cardinalityFormatter(cardinality)}
      </span>
    </BarAndLabel>
    <TooltipContent slot="tooltip-content">
      {formatInteger(cardinality)} unique values
    </TooltipContent>
  </Tooltip>
{/if}
