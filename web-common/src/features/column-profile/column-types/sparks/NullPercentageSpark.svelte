<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-common/layout/config";
  import {
    DATA_TYPE_COLORS,
    isNested,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import { singleDigitPercentage } from "@rilldata/web-common/lib/formatters";
  import BarAndLabel from "../../../../components/BarAndLabel.svelte";

  export let type: string;
  export let nullCount: number;
  export let totalRows: number;

  let percentage: number;
  $: if (nullCount !== undefined && totalRows !== undefined)
    percentage = nullCount / totalRows;
  $: innerType = isNested(type) ? "STRUCT" : type;
</script>

{#if totalRows !== undefined && nullCount !== undefined && !isNaN(percentage) && percentage <= 1}
  <Tooltip location="right" distance={8}>
    <BarAndLabel
      compact
      showBackground={nullCount !== 0}
      color={DATA_TYPE_COLORS[innerType]?.bgClass}
      value={percentage || 0}
    >
      <span
        style:font-size="{COLUMN_PROFILE_CONFIG.fontSize}px"
        class="ui-copy-number"
        class:text-gray-300={nullCount === 0}
        >{singleDigitPercentage(percentage)}</span
      >
    </BarAndLabel>
    <TooltipContent slot="tooltip-content">
      {#if nullCount > 0}
        {singleDigitPercentage(percentage)} of the values are null
      {:else}
        no null values in this column
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
