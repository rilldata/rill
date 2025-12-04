<script lang="ts" context="module">
  import { writable } from "svelte/store";
  const columnLengths = writable(new Map<string, number>());
</script>

<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type { MeasureColumnProps } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import {
    calculateColumnWidth,
    calculateMeasureWidth,
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { cellInspectorStore } from "../stores/cell-inspector-store";
  import type { Cell, Column, HeaderGroup, Row } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import type { PivotDataRow } from "./types";

  // State props
  export let assembled: boolean;
  export let measures: MeasureColumnProps;
  export let dataRows: PivotDataRow[];
  export let hasMeasureContextColumns: boolean;
  export let canShowDataViewer = false;
  export let activeCell: { rowId: string; columnId: string } | null | undefined;

  // Table props
  export let headerGroups: HeaderGroup<PivotDataRow>[];
  export let rows: Row<PivotDataRow>[];
  export let virtualRows: { index: number }[];
  export let totalsRow: PivotDataRow | undefined;
  export let before: number;
  export let after: number;
  export let totalRowSize: number;

  // Event handlers
  export let onCellClick: (e: MouseEvent) => void;
  export let onMouseMove: (e: MouseEvent) => void;
  export let onTableLeave: () => void;
  export let onCellCopy: (e: MouseEvent) => void;

  const HEADER_HEIGHT = 30;

  let totalLength = 0;

  $: headers = headerGroups[0].headers;

  // Initialize column lengths if not already set
  $: headers.forEach((header) => {
    const columnId = header.column.id;

    if (!$columnLengths.has(columnId)) {
      const measure = getMeasureColumn(header.column);
      const estimatedWidth = measure
        ? calculateMeasureWidth(
            measure.name,
            measure.label,
            measure.formatter,
            totalsRow,
            dataRows,
          )
        : calculateColumnWidth(
            String(header.column.columnDef.header),
            "",
            dataRows,
          );
      columnLengths.update((lengths) => lengths.set(columnId, estimatedWidth));
    }
  });

  $: totalLength = headers.reduce((acc, header) => {
    return (
      acc + ($columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH)
    );
  }, 0);

  function getMeasureColumn(headerColumn: Column<PivotDataRow>) {
    const columnId = headerColumn.id;
    return measures.find((m) => m.name === columnId);
  }

  function isCellActive(cell: Cell<PivotDataRow, unknown>) {
    return (
      cell.row.id === activeCell?.rowId &&
      cell.column.id === activeCell?.columnId
    );
  }

  function hasBorderRight(columnId: string): boolean {
    if (!hasMeasureContextColumns) return true;
    const measureIndex = measures.findIndex((m) => m.name === columnId);
    if (measureIndex === -1) return true;
    //  Every third column is the last in its group
    return (measureIndex + 1) % 3 === 0;
  }
</script>

<div
  class="w-full absolute top-0 z-50 flex pointer-events-none"
  style:width="{totalLength}px"
  style:height="{totalRowSize + HEADER_HEIGHT + headerGroups.length}px"
>
  {#each headers as header, i (header.id)}
    {@const length =
      $columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH}
    {@const last = i === headers.length - 1}
    <div style:width="{length}px" class="h-full relative">
      <Resizer
        side="right"
        direction="EW"
        min={WIDTHS.MIN_MEASURE_WIDTH}
        max={WIDTHS.MAX_MEASURE_WIDTH}
        dimension={length}
        justify={last ? "end" : "center"}
        hang={!last}
        onUpdate={(d) =>
          columnLengths.update((lengths) => {
            return lengths.set(header.column.id, d);
          })}
      >
        <div class="resize-bar" />
      </Resizer>
    </div>
  {/each}
</div>

<table
  role="presentation"
  style:width="{totalLength}px"
  class:with-measure={measures.length > 0}
  on:click={modified({ shift: onCellCopy, click: onCellClick })}
  on:mousemove={onMouseMove}
  on:mouseleave={onTableLeave}
>
  <colgroup>
    {#each headers as header (header.id)}
      {@const length =
        $columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH}
      <col style:width="{length}px" style:max-width="{length}px" />
    {/each}
  </colgroup>

  <thead>
    {#each headerGroups as headerGroup (headerGroup.id)}
      <tr>
        {#each headerGroup.headers as header (header.id)}
          {@const sortDirection = header.column.getIsSorted()}
          {@const icon = header.column.columnDef.meta?.icon}
          <th>
            <button
              class="header-cell"
              class:cursor-pointer={header.column.getCanSort()}
              class:select-none={header.column.getCanSort()}
              class:flex-row-reverse={!!getMeasureColumn(header.column)}
              class:border-r={hasBorderRight(header.column.id)}
              on:click={header.column.getToggleSortingHandler()}
            >
              {#if !header.isPlaceholder}
                {#if icon}
                  <svelte:component this={icon} />
                {:else}
                  <p class="truncate">
                    {header.column.columnDef.header}
                  </p>
                {/if}
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
        {#each cells as cell (cell.id)}
          {@const result =
            typeof cell.column.columnDef.cell === "function"
              ? cell.column.columnDef.cell(cell.getContext())
              : cell.column.columnDef.cell}
          {@const isActive = isCellActive(cell)}
          <td
            class="ui-copy-number cell truncate"
            class:active-cell={isActive}
            class:interactive-cell={canShowDataViewer &&
              cell.getValue() !== undefined}
            class:text-right={getMeasureColumn(cell.column)}
            class:border-r={hasBorderRight(cell.column.id)}
            class:total-label={cell.getValue() === "Total"}
            data-value={cell.getValue()}
            data-rowid={cell.row.id}
            data-columnid={cell.column.id}
            on:mouseover={() => {
              const value = cell.getValue();
              if (value !== undefined && value !== null) {
                // Always update the value in the store, but don't change visibility
                cellInspectorStore.updateValue(String(value));
              }
            }}
            on:focus={() => {
              const value = cell.getValue();
              if (value !== undefined && value !== null) {
                // Always update the value in the store, but don't change visibility
                cellInspectorStore.updateValue(String(value));
              }
            }}
          >
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
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            {/if}
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

  .resize-bar {
    @apply bg-primary-500 w-1 h-full;
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
    @apply border-b relative;
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
    @apply p-0 m-0;
  }

  .header-cell {
    @apply px-2 bg-surface size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    @apply text-gray-800 font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2 text-gray-800;
  }

  tr > td {
    @apply font-normal;
  }

  /* The totals row */
  .with-measure tbody > tr:nth-of-type(2) {
    @apply bg-surface sticky z-20;
    top: var(--total-header-height);
  }

  /* The totals row label - make it bold for flat tables */
  .total-label {
    @apply font-semibold;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }

  tr:hover .active-cell {
    @apply bg-primary-100;
  }

  .interactive-cell {
    @apply cursor-pointer;
  }
  .interactive-cell.cell:hover {
    @apply bg-primary-100;
  }
  .active-cell.cell {
    @apply bg-primary-50;
  }
</style>
