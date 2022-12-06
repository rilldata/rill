<script lang="ts">
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import { singleDigitPercentage } from "@rilldata/web-local/lib/util/formatters";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let type: string;
  export let nullCount: number;
  export let totalRows: number;

  $: percentage = nullCount / totalRows;
</script>

{#if totalRows !== undefined && nullCount !== undefined && !isNaN(percentage)}
  <Tooltip location="right" alignment="center" distance={8}>
    <BarAndLabel
      showBackground={nullCount !== 0}
      color={DATA_TYPE_COLORS[type]?.bgClass}
      value={percentage || 0}
    >
      <span class:text-gray-300={nullCount === 0}
        >∅ {singleDigitPercentage(percentage)}</span
      >
    </BarAndLabel>
    <TooltipContent slot="tooltip-content">
      <svelte:fragment slot="title">
        what percentage of values are null?
      </svelte:fragment>
      {#if nullCount > 0}
        {singleDigitPercentage(percentage)} of the values are null
      {:else}
        no null values in this column
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
