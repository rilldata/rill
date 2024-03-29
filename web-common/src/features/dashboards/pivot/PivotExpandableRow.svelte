<script lang="ts">
  import type { Row, Cell } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import { pivotFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";

  export let row: Row<PivotDataRow>;
  export let measureCount: number;
  export let hasDimension: boolean;
  export let totals = false;

  $: cells = row.getVisibleCells() as Cell<
    PivotDataRow,
    string | number | undefined
  >[];
  $: rowHeader = cells[0].getValue() as string | null;
  $: canExpand = row.getCanExpand();
  $: expanded = row.getIsExpanded();
</script>

<tr class:with-row-dimension={hasDimension} class:totals>
  {#if hasDimension}
    <td class="row-header">
      <svelte:element
        this={canExpand ? "button" : "div"}
        role={canExpand ? "button" : undefined}
        class="cell gap-x-1 flex items-center w-full"
        class:font-normal={row.depth >= 1}
        on:click={canExpand ? () => row.toggleExpanded() : undefined}
      >
        {#if rowHeader === "LOADING_CELL"}
          <span class="loading-cell" />
        {:else}
          <div
            class="gap-x-1 flex items-center w-full"
            style:padding-left="{row.depth * 1.5}rem"
          >
            {#if canExpand}
              <span class="transition-transform" class:rotate-90={expanded}>
                <ChevronRight />
              </span>
            {/if}
            {rowHeader}
          </div>
        {/if}
      </svelte:element>
    </td>
  {/if}

  {#each hasDimension ? cells.slice(1) : cells as cell, i (cell.id)}
    {@const value = cell.getValue()}
    {@const format = String(cell.column.columnDef.cell)}
    <td class:border-r={i && (i + 1) % measureCount === 0}>
      <div class="cell ui-copy-number">
        {#if value === undefined || value === null}
          <span class="no-data">no data</span>
        {:else}
          {pivotFormatter(value, format)}
        {/if}
      </div>
    </td>
  {/each}
</tr>

<style lang="postcss">
  * {
    @apply border-slate-200;
  }

  .no-data {
    @apply italic text-gray-400;
    font-size: 0.925em;
  }

  td {
    @apply p-0 m-0 text-xs text-right;
    height: var(--row-height);
  }

  .row-header:not(:last-of-type) {
    @apply border-r font-medium;
  }

  .cell {
    @apply p-1 px-2 truncate;
  }

  .totals td:first-of-type > .cell {
    @apply font-bold;
  }

  td:last-of-type {
    @apply border-r-0;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }

  .loading-cell {
    @apply h-2 mx-2 bg-gray-200 rounded w-full inline-block;
  }

  .row-header {
    @apply sticky left-0 z-10;
    @apply text-left bg-white;
  }

  .totals {
    @apply bg-slate-100 font-bold;
  }
</style>
