<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { Row } from "tanstack-table-8-svelte-5";

  export let row: Row<PivotDataRow>;
  export let dimensionLabel: string;
  export let currentLimit: number;
  // Undefined past the last limit step: the row then only notes that the
  // rest of the values are left out.
  export let nextLimit: number | undefined;
  export let hasNestedDimensions = false;

  $: needsSpacer = row.depth >= 1 || hasNestedDimensions;
</script>

<div class="show-more-cell" style:padding-left="{row?.depth * 14}px">
  {#if needsSpacer}
    <Spacer size="16px" />
  {/if}
  <Tooltip distance={8} location="right">
    {#if nextLimit}
      <span class="show-more-link">{m.pivot_show_more()}</span>
    {:else}
      <span class="text-fg-secondary italic">
        {m.pivot_rows_truncated({ limit: currentLimit })}
      </span>
    {/if}
    <TooltipContent slot="tooltip-content">
      {nextLimit
        ? m.pivot_show_more_tooltip({
            limit: nextLimit,
            dimension: dimensionLabel,
          })
        : m.pivot_rows_truncated_tooltip({
            limit: currentLimit,
            dimension: dimensionLabel,
          })}
    </TooltipContent>
  </Tooltip>
</div>

<style lang="postcss">
  .show-more-cell {
    @apply flex items-center gap-x-0.5 h-full;
  }

  .show-more-link {
    @apply text-primary-600 cursor-pointer;
  }
</style>
