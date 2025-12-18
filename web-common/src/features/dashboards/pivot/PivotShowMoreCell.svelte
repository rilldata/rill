<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { Row } from "@tanstack/svelte-table";

  export let row: Row<PivotDataRow>;
  export let value: string;
  export let assembled = true;
</script>

<div class="show-more-cell" style:padding-left="{row?.depth * 14}px">
  <Spacer size="16px" />
  <span class="italic {assembled ? 'ui-copy' : 'ui-copy-inactive'}">
    {value}
  </span>
</div>

<style lang="postcss">
  .show-more-cell {
    @apply flex items-center gap-x-0.5 h-full;
  }

  .show-more-cell:hover span {
    @apply text-primary-600;
  }
</style>
