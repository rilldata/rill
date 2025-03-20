<script lang="ts">
  import type { Row } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";
  import PivotExpandableCell from "./PivotExpandableCell.svelte";

  export let rawValue: string | number | null | undefined;
  export let formattedValue: string | number | undefined;
  export let columnId: string | undefined = undefined;
  export let rowId: string | undefined = undefined;

  export let showRightBorder = false;
  export let comparisonType = false;
  export let active = false;
  export let interactive = false;

  export let assembled = true;
  export let row: Row<PivotDataRow> | null = null;
</script>

<td
  data-rowid={rowId}
  data-columnid={columnId}
  data-value={rawValue}
  class="ui-copy-number"
  class:active
  class:interactive
  class:comparison-value={comparisonType}
  class:border-r={showRightBorder}
  class:is-null={rawValue === undefined || rawValue == null}
  class:negative={typeof rawValue === "number" && rawValue < 0}
>
  {#if row}
    <PivotExpandableCell {row} value={rawValue} {assembled} />
  {:else if assembled}
    {formattedValue ?? rawValue ?? "-"}
  {:else}
    <span class="loading-cell" />
  {/if}
</td>

<style lang="postcss">
  td {
    @apply m-0 size-full p-1 px-2;
    @apply text-right truncate;
    @apply whitespace-nowrap text-xs text-gray-800;
    height: var(--row-height);
  }

  .comparison-value {
    @apply text-gray-500;
  }

  .negative {
    @apply text-red-500;
  }

  .is-null {
    @apply text-gray-400;
  }

  td:last-of-type {
    @apply border-r-0;
  }

  :global(.with-row-dimension tr) > td:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  :global(tbody > tr:nth-of-type(2)) > td:first-of-type {
    @apply font-semibold;
  }

  :global(tbody tr:hover) td:first-of-type {
    @apply bg-slate-100;
  }

  .active {
    @apply bg-primary-50;
  }

  :global(tr:hover) .active {
    @apply bg-primary-100;
  }

  .interactive {
    @apply cursor-pointer;
  }

  .interactive:hover {
    @apply bg-primary-100;
  }

  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
