<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { Row } from "@tanstack/svelte-table";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
  export let hasNestedDimensions = false;

  $: needsSpacer = row.depth >= 1 || hasNestedDimensions;
</script>

<div class="show-more-cell" style:padding-left="{row?.depth * 14}px">
  {#if needsSpacer}
    <Spacer size="16px" />
  {/if}
  <Tooltip distance={8} location="right">
    <span class={assembled ? "text-fg-primary" : "text-fg-disabled"}>
      Show more ...
    </span>
    <TooltipContent slot="tooltip-content">
      {value}
    </TooltipContent>
  </Tooltip>
</div>

<style lang="postcss">
  .show-more-cell {
    @apply flex items-center gap-x-0.5 h-full;
  }

  .show-more-cell span {
    @apply text-primary-600;
  }
</style>
