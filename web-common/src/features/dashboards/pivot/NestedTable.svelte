<script lang="ts" context="module">
  import { writable } from "svelte/store";
  const measureLengths = writable(new Map<string, number>());
</script>

<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import {
    calculateColumnWidth,
    calculateMeasureWidth,
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import type { Cell, HeaderGroup, Row } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import type { MeasureColumnProps } from "./pivot-column-definition";
  import type { PivotDataRow } from "./types";

  // State props
  export let hasRowDimension: boolean;
  export let hasColumnDimension: boolean;
  export let timeDimension: string;
  export let rowDimensionLabel: string;
  export let assembled: boolean;
  export let dataRows: PivotDataRow[];
  export let measures: MeasureColumnProps;
  export let totalsRow: PivotDataRow | undefined;
  export let canShowDataViewer = false;
  export let activeCell: { rowId: string; columnId: string } | null | undefined;

  // Table props
  export let headerGroups: HeaderGroup<PivotDataRow>[];
  export let rows: Row<PivotDataRow>[];
  export let virtualRows: { index: number }[];
  export let after: number;
  export let before: number;
  export let containerRefElement: HTMLDivElement;
  export let scrollLeft: number;
  export let totalRowSize: number;

  // Event handlers
  export let onCellClick: (cell: Cell<PivotDataRow, unknown>) => void;
  export let onCellHover: (
    e: MouseEvent & { currentTarget: EventTarget & HTMLElement },
  ) => void;
  export let onCellLeave: () => void;
  export let onCellCopy: (e: MouseEvent) => void;

  const HEADER_HEIGHT = 30;

  let resizingMeasure = false;
  let initialMeasureIndexOnResize = 0;
  let initLengthOnResize = 0;
  let initScrollOnResize = 0;
  let percentOfChangeDuringResize = 0;

  $: rowDimensionName = rowDimensionLabel ? rowDimensionLabel : null;

  $: rowDimensionWidth =
    hasRowDimension && rowDimensionName
      ? calculateColumnWidth(rowDimensionName, timeDimension, dataRows)
      : 0;

  $: measures.forEach(({ name, label, formatter }) => {
    if (!$measureLengths.has(name)) {
      const estimatedWidth = calculateMeasureWidth(
        name,
        label,
        formatter,
        totalsRow,
        dataRows,
      );

      measureLengths.update((measureLengths) => {
        return measureLengths.set(name, estimatedWidth);
      });
    }
  });

  $: if (resizingMeasure && containerRefElement && measureLengths) {
    containerRefElement.scrollTo({
      left:
        initScrollOnResize +
        percentOfChangeDuringResize * (totalLength - initLengthOnResize),
    });
  }

  $: measureCount = measures.length;

  $: subHeaders = [
    {
      subHeaders: measures.map((m) => ({
        column: { columnDef: { name: m.name } },
      })),
    },
  ];

  let measureGroups: {
    subHeaders: { column: { columnDef: { name: string } } }[];
  }[];
  // @ts-expect-error - I have manually added the name property in pivot-column-definition.ts
  $: measureGroups =
    headerGroups[headerGroups.length - 2]?.headers?.slice(
      hasRowDimension ? 1 : 0,
    ) ?? subHeaders;

  $: measureGroupsLength = measureGroups.length;
  $: totalMeasureWidth = measures.reduce(
    (acc, { name }) => acc + ($measureLengths.get(name) ?? 0),
    0,
  );
  $: totalLength = measureGroupsLength * totalMeasureWidth;

  function isMeasureColumn(header, colNumber: number) {
    // Measure columns are the last columns in the header group
    if (header.depth !== headerGroups.length) return;
    // If there is a row dimension, the first column is not a measure column
    if (!rowDimensionName) {
      return true;
    } else return colNumber > 0;
  }

  function onResizeStart(e: MouseEvent) {
    initLengthOnResize = totalLength;
    initScrollOnResize = scrollLeft;

    const offset =
      e.clientX -
      containerRefElement.getBoundingClientRect().left -
      rowDimensionWidth -
      measures.reduce((rollingSum, { name }, i) => {
        return i <= initialMeasureIndexOnResize
          ? rollingSum + ($measureLengths.get(name) ?? 0)
          : rollingSum;
      }, 0) +
      4;

    percentOfChangeDuringResize = (scrollLeft + offset) / totalLength;
  }

  function isCellActive(cell: Cell<PivotDataRow, unknown>) {
    return (
      cell.row.id === activeCell?.rowId &&
      cell.column.id === activeCell?.columnId
    );
  }

  function shouldShowHeaderRightBorder(header: any, index: number): boolean {
    const isMeasure = isMeasureColumn(header, index);
    if (!isMeasure) return true;

    let offset = 0;
    if (!hasRowDimension) offset = 1;
    return (index + offset) % measureCount === 0 && index > 0;
  }

  function shouldShowRightBorder(index: number): boolean {
    let offset = 0;
    if (!hasRowDimension) offset = 1;
    return (index + offset) % measureCount === 0;
  }

  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;
</script>

<div
  class="w-full absolute top-0 z-50 flex pointer-events-none"
  style:width="{totalLength + rowDimensionWidth}px"
  style:height="{totalRowSize + totalHeaderHeight + headerGroups.length}px"
>
  <div style:width="{rowDimensionWidth}px" class="sticky left-0 flex-none flex">
    <Resizer
      side="right"
      direction="EW"
      min={WIDTHS.MIN_COL_WIDTH}
      max={WIDTHS.MAX_COL_WIDTH}
      dimension={rowDimensionWidth}
      onUpdate={(d) => (rowDimensionWidth = d)}
      onMouseDown={(e) => {
        resizingMeasure = false;
        onResizeStart(e);
      }}
      onMouseUp={() => {
        resizingMeasure = false;
      }}
    >
      <div class="resize-bar" />
    </Resizer>
  </div>

  {#each measureGroups as { subHeaders }, groupIndex (groupIndex)}
    <div class="h-full z-50 flex" style:width="{totalMeasureWidth}px">
      {#each subHeaders as { column: { columnDef: { name } } }, i (name)}
        {@const length = $measureLengths.get(name) ?? WIDTHS.INIT_MEASURE_WIDTH}
        {@const last =
          i === subHeaders.length - 1 &&
          groupIndex === measureGroups.length - 1}
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
              measureLengths.update((measureLengths) => {
                return measureLengths.set(name, d);
              })}
            onMouseDown={(e) => {
              resizingMeasure = true;
              onResizeStart(e);
            }}
            onMouseUp={() => {
              resizingMeasure = false;
            }}
          >
            <div class="resize-bar" />
          </Resizer>
        </div>
      {/each}
    </div>
  {/each}
</div>

<table
  class:with-row-dimension={hasRowDimension}
  class:with-col-dimension={hasColumnDimension}
  role="presentation"
  style:width="{totalLength + rowDimensionWidth}px"
  on:click={modified({ shift: onCellCopy })}
>
  <colgroup>
    {#if rowDimensionName && rowDimensionWidth}
      <col
        style:width="{rowDimensionWidth}px"
        style:max-width="{rowDimensionWidth}px"
      />
    {/if}

    {#each measureGroups as { subHeaders }, i (i)}
      {#each subHeaders as { column: { columnDef: { name } } } (name)}
        {@const length = $measureLengths.get(name)}
        <col style:width="{length}px" style:max-width="{length}px" />
      {/each}
    {/each}
  </colgroup>

  <thead>
    {#each headerGroups as headerGroup (headerGroup.id)}
      <tr>
        {#each headerGroup.headers as header, i (header.id)}
          {@const sortDirection = header.column.getIsSorted()}
          {@const icon = header.column.columnDef.meta?.icon}

          <th colSpan={header.colSpan}>
            <button
              class="header-cell"
              class:cursor-pointer={header.column.getCanSort()}
              class:select-none={header.column.getCanSort()}
              class:flex-row-reverse={isMeasureColumn(header, i)}
              class:border-r={shouldShowHeaderRightBorder(header, i)}
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
            class:border-r={shouldShowRightBorder(i)}
            on:click={() => onCellClick(cell)}
            on:mouseenter={onCellHover}
            on:mouseleave={onCellLeave}
            data-value={cell.getValue()}
            class:totals-column={i > 0 && i <= measureCount}
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

  .with-row-dimension thead tr th:first-of-type .header-cell {
    @apply flex-row-reverse;
  }

  .with-row-dimension thead tr:last-of-type th:first-of-type .header-cell {
    @apply flex-row;
  }

  tbody .cell {
    height: var(--row-height);
  }

  th {
    @apply p-0 m-0 text-xs;
    @apply border-b relative;
  }

  /* thead tr:last-of-type th {
    @apply border-b;
  } */

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
    @apply text-gray-800 font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2 text-gray-800;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  .with-row-dimension thead > tr:not(:last-of-type) > th:first-of-type {
    @apply border-b-0;
  }

  .with-row-dimension tr > th:first-of-type {
    @apply sticky left-0 z-20;
    @apply bg-white;
  }

  .with-row-dimension tr > td:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  /* tr > td:first-of-type:not(:last-of-type) {
    @apply border-r;
  } */

  /* The totals row */
  tbody > tr:nth-of-type(2) {
    @apply bg-white sticky z-20;
    top: var(--total-header-height);
    height: calc(var(--row-height) + 2px);
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
