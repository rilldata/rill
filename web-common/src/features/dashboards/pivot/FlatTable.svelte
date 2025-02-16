<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import type { Cell, HeaderGroup, Row } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  export let headerGroups: HeaderGroup<PivotDataRow>[];
  export let rows: Row<PivotDataRow>[];
  export let virtualRows: { index: number }[];
  export let before: number;
  export let after: number;
  export let canShowDataViewer = false;
  export let activeCell: { rowId: string; columnId: string } | null | undefined;
  export let onCellClick: (cell: Cell<PivotDataRow, unknown>) => void;
  export let onCellHover: (
    e: MouseEvent & { currentTarget: EventTarget & HTMLElement },
  ) => void;
  export let onCellLeave: () => void;
  export let onCellCopy: (e: MouseEvent) => void;
  export let assembled: boolean;

  function isMeasureColumn(header, colNumber: number) {
    // For flat tables, measures are always at the end after dimensions
    // We can identify them by checking if they are leaf columns (no subcolumns)
    // and if their column definition has an accessorKey (which measures have)
    return !header.column.columns && header.column.columnDef.accessorKey;
  }

  function isCellActive(cell: Cell<PivotDataRow, unknown>) {
    return (
      cell.row.id === activeCell?.rowId &&
      cell.column.id === activeCell?.columnId
    );
  }
</script>

<table role="presentation" on:click={modified({ shift: onCellCopy })}>
  <thead>
    {#each headerGroups as headerGroup (headerGroup.id)}
      <tr>
        {#each headerGroup.headers as header, i (header.id)}
          {@const sortDirection = header.column.getIsSorted()}

          <th>
            <button
              class="header-cell"
              class:cursor-pointer={header.column.getCanSort()}
              class:select-none={header.column.getCanSort()}
              class:flex-row-reverse={isMeasureColumn(header, i)}
              on:click={header.column.getToggleSortingHandler()}
            >
              {#if !header.isPlaceholder}
                <p class="truncate">
                  {header.column.columnDef.header}
                </p>
                {#if sortDirection}
                  <span
                    class="transition-transform -mr-1"
                    class:-rotate-180={sortDirection === "asc"}
                  >
                    <ArrowDown />
                  </span>
                {/if}
              {/if}
            </button>
          </th>
        {/each}
      </tr>
    {/each}
  </thead>
  <tbody>
    <tr style:height="{before}px" />
    {#each virtualRows as row (row.index)}
      {@const cells = rows[row.index].getVisibleCells()}
      <tr>
        {#each cells as cell, i (cell.id)}
          {@const result =
            typeof cell.column.columnDef.cell === "function"
              ? cell.column.columnDef.cell(cell.getContext())
              : cell.column.columnDef.cell}
          {@const isActive = isCellActive(cell)}
          <td
            class="ui-copy-number"
            class:active-cell={isActive}
            class:interactive-cell={canShowDataViewer}
            on:click={() => onCellClick(cell)}
            on:mouseenter={onCellHover}
            on:mouseleave={onCellLeave}
            data-value={cell.getValue()}
          >
            <div class="cell pointer-events-none truncate" role="presentation">
              {#if result?.component && result?.props}
                <svelte:component
                  this={result.component}
                  {...result.props}
                  {assembled}
                />
              {:else if typeof result === "string" || typeof result === "number"}
                {result}
              {:else}
                <svelte:component
                  this={flexRender(
                    cell.column.columnDef.cell,
                    cell.getContext(),
                  )}
                />
              {/if}
            </div>
          </td>
        {/each}
      </tr>
    {/each}
    <tr style:height="{after}px" />
  </tbody>
</table>

<style lang="postcss">
  * {
    @apply border-slate-200;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal;
    @apply bg-surface table-fixed;
  }

  /* Pin header */
  thead {
    @apply sticky top-0;
    @apply z-30 bg-surface;
  }

  tbody .cell {
    height: var(--row-height);
  }

  th {
    @apply p-0 m-0 text-xs;
    @apply border-r border-b relative;
  }

  th:last-of-type,
  td:last-of-type {
    @apply border-r-0;
  }

  th,
  td {
    @apply whitespace-nowrap text-xs;
  }

  td {
    @apply text-right;
    @apply p-0 m-0;
  }

  .header-cell {
    @apply px-2 bg-white size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    @apply font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2;
  }

  tr > td:first-of-type:not(:last-of-type) {
    @apply border-r font-normal;
  }

  /* The totals row */
  tbody > tr:nth-of-type(2) {
    @apply bg-slate-50 sticky z-20 font-semibold;
    top: var(--total-header-height);
  }

  /* The totals row header */
  tbody > tr:nth-of-type(2) > td:first-of-type {
    @apply font-semibold;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }

  tr:hover .active-cell .cell {
    @apply bg-primary-100;
  }

  .interactive-cell {
    @apply cursor-pointer;
  }
  .interactive-cell:hover .cell {
    @apply bg-primary-100;
  }
  .active-cell .cell {
    @apply bg-primary-50;
  }
</style>
