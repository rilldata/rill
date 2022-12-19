<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { DATA_TYPE_COLORS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { singleDigitPercentage } from "@rilldata/web-common/lib/formatters";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import BarAndLabel from "../../../viz/BarAndLabel.svelte";

  export let type: string;
  export let nullCount: number;
  export let totalRows: number;

  $: percentage = nullCount / totalRows;
</script>

{#if totalRows !== undefined && nullCount !== undefined && !isNaN(percentage)}
  <Tooltip location="right" alignment="center" distance={8}>
    <BarAndLabel
      compact
      showBackground={nullCount !== 0}
      color={DATA_TYPE_COLORS[type]?.bgClass}
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
