<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { Row } from "@tanstack/svelte-table";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
</script>

<div class="show-more-cell" style:padding-left="{row?.depth * 14}px">
  <Spacer size="16px" />
  <Tooltip distance={8} location="right">
    <span class="text-primary-600 {assembled ? 'ui-copy' : 'ui-copy-inactive'}">
      Show more ...
    </span>
    <TooltipContent slot="tooltip-content">
      {value}
    </TooltipContent>
  </Tooltip>
</div>

<style lang="postcss">
  .show-more-cell {
    @apply flex items-center h-full;
  }

  .show-more-cell span {
    @apply text-primary-600;
  }
</style>
