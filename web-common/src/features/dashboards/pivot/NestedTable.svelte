<script lang="ts" context="module">
  import { writable } from "svelte/store";
  const measureLengths = writable(new Map<string, number>());
</script>

<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import type { Cell, HeaderGroup, Row } from "tanstack-table-8-svelte-5";
  import { flexRender } from "tanstack-table-8-svelte-5";
  import { cellInspectorStore } from "../stores/cell-inspector-store";
  import type { PivotClickSelectionState } from "./pivot-click-selection";
  import {
    getRowNestedLabel,
    type DimensionColumnProps,
    type MeasureColumnProps,
  } from "./pivot-column-definition";
  import {
    calculateMeasureWidth,
    calculateRowDimensionWidth,
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
  } from "./pivot-column-width-utils";
  import type { PivotRowSelectionState } from "./pivot-row-selection";
  import { isShowMoreRow } from "./pivot-utils";
  import type { PivotDataRow } from "./types";

  // State props
  export let hasColumnDimension: boolean;
  export let timeDimension: string;
  export let assembled: boolean;
  export let rowDimensions: DimensionColumnProps;
  export let dataRows: PivotDataRow[];
  export let measures: MeasureColumnProps;
  export let totalsRow: PivotDataRow | undefined;
  export let canShowDataViewer = false;
  export let enableClickToFilter = false;
  export let rowSelectionState: PivotRowSelectionState | undefined = undefined;
  export let clickSelection: PivotClickSelectionState | undefined = undefined;
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
  export let onMouseMove: (e: MouseEvent) => void;
  export let onCellClick: (e: MouseEvent) => void;
  export let onTableLeave: () => void;
  export let onCellCopy: (e: MouseEvent) => void;
  export let onColumnHeaderClick:
    | ((dimensionPath: Record<string, string>) => void)
    | undefined = undefined;

  const HEADER_HEIGHT = 30;

  // Hover state for column dimension headers
  let hoveredColRange: { start: number; size: number } | null = null;

  // Compute which leaf column indices are covered by selected column headers.
  // Iterates header groups to find selected headers and collects their ranges.
  $: selectedColIndices = (() => {
    if (!clickSelection?.hasAnySelection) return new Set<number>();
    const indices = new Set<number>();
    for (const group of headerGroups) {
      let colStart = 0;
      for (const header of group.headers) {
        const meta = header.column.columnDef.meta;
        if (
          meta?.dimensionPath &&
          clickSelection.isColumnHeaderSelected(meta.dimensionPath)
        ) {
          for (let c = colStart; c < colStart + header.colSpan; c++) {
            indices.add(c);
          }
        }
        colStart += header.colSpan;
      }
    }
    return indices;
  })();

  // Compute which leaf column indices have a cell selected via click-to-filter.
  // Uses the last header group (leaf columns) to map column IDs to indices.
  $: cellSelectedColIndices = (() => {
    if (!clickSelection?.selectedCellColumnIds?.size) return new Set<number>();
    const leafGroup = headerGroups[headerGroups.length - 1];
    if (!leafGroup) return new Set<number>();
    const indices = new Set<number>();
    let colIdx = 0;
    for (const header of leafGroup.headers) {
      if (clickSelection.selectedCellColumnIds.has(header.column.id)) {
        indices.add(colIdx);
      }
      colIdx += header.colSpan;
    }
    return indices;
  })();

  // Check if a header (by its leaf column range) contains any cell-selected columns
  function isInCellSelectedColRange(
    colStart: number,
    colSpan: number,
  ): boolean {
    if (cellSelectedColIndices.size === 0) return false;
    for (let i = colStart; i < colStart + colSpan; i++) {
      if (cellSelectedColIndices.has(i)) return true;
    }
    return false;
  }

  let resizingMeasure = false;
  let initialMeasureIndexOnResize = 0;
  let initLengthOnResize = 0;
  let initScrollOnResize = 0;
  let percentOfChangeDuringResize = 0;

  $: hasRowDimension = rowDimensions.length > 0;
  $: hasExpandableRows = rowDimensions.length > 1;
  $: hasMeasures = measures.length > 0;
  $: rowDimensionLabel = getRowNestedLabel(rowDimensions);
  $: rowDimensionName = rowDimensionLabel ? rowDimensionLabel : null;

  $: rowDimensionWidth =
    hasRowDimension && rowDimensionName
      ? calculateRowDimensionWidth(rowDimensionName, timeDimension, dataRows)
      : 0;

  $: {
    // Get the longest column dimension header to ensure proper width calculation
    const maxColumnDimensionHeader = getMaxColumnDimensionHeader(
      hasColumnDimension,
      headerGroups,
    );

    measures.forEach(({ name, label, formatter }) => {
      if (!$measureLengths.has(name)) {
        const estimatedWidth = calculateMeasureWidth(
          name,
          label,
          formatter,
          totalsRow,
          dataRows,
          hasColumnDimension ? maxColumnDimensionHeader : undefined,
        );

        measureLengths.update((measureLengths) => {
          return measureLengths.set(name, estimatedWidth);
        });
      }
    });
  }

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

  function isCellClicked(cell: Cell<PivotDataRow, unknown>) {
    return clickSelection?.isCellSelected(cell.row.id, cell.column.id) ?? false;
  }

  function isHeaderInHoveredRange(
    headerStart: number,
    headerSize: number,
  ): boolean {
    if (!hoveredColRange) return false;
    const hovEnd = hoveredColRange.start + hoveredColRange.size;
    return (
      headerStart >= hoveredColRange.start && headerStart + headerSize <= hovEnd
    );
  }

  function isHoveredHeader(colStart: number, colSpan: number): boolean {
    if (!hoveredColRange) return false;
    return (
      colStart === hoveredColRange.start && colSpan === hoveredColRange.size
    );
  }

  // Returns true if this header should be highlighted as falling within a
  // selected column range (i.e. it is a child of the clicked header).
  function isInSelectedColRange(
    colStart: number,
    colSpan: number,
    isSelfSelected: boolean,
  ): boolean {
    if (selectedColIndices.size === 0 || colSpan === 0 || isSelfSelected) {
      return false;
    }
    for (let i = colStart; i < colStart + colSpan; i++) {
      if (!selectedColIndices.has(i)) return false;
    }
    return true;
  }

  function shouldShowHeaderRightBorder(header: any, index: number): boolean {
    const isMeasure = isMeasureColumn(header, index);
    if (!isMeasure) return true;

    let offset = 0;
    if (!hasRowDimension) offset = 1;
    return (index + offset) % measureCount === 0 && index > 0;
  }

  function getMaxColumnDimensionHeader(
    hasColumnDimension: boolean,
    headerGroups: HeaderGroup<PivotDataRow>[],
  ): string {
    if (!hasColumnDimension || headerGroups.length === 0) return "";

    // Get the second-to-last header group which contains column dimension values
    const colDimensionHeaderGroup =
      headerGroups.length >= 2
        ? headerGroups[headerGroups.length - 2]
        : undefined;
    if (!colDimensionHeaderGroup?.headers) return "";

    return colDimensionHeaderGroup.headers.reduce((longest, header) => {
      const headerText = String(header.column?.columnDef?.header ?? "");
      return headerText.length > longest.length ? headerText : longest;
    }, "");
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
      <div class="resize-bar"></div>
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
            <div class="resize-bar"></div>
          </Resizer>
        </div>
      {/each}
    </div>
  {/each}
</div>

<table
  class:with-row-dimension={hasRowDimension}
  class:with-col-dimension={hasColumnDimension}
  class:with-expandable-rows={hasExpandableRows}
  class:with-totals-row={!!totalsRow}
  class:with-measures={hasMeasures}
  role="presentation"
  style:width="{totalLength + rowDimensionWidth}px"
  onclick={modified({ shift: onCellCopy, click: onCellClick })}
  onmousemove={onMouseMove}
  onmouseleave={() => {
    hoveredColRange = null;
    onTableLeave();
  }}
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
          {@const dimMeta = header.column.columnDef.meta}
          {@const icon = dimMeta?.icon}
          {@const isColDimHeader =
            !header.isPlaceholder && !!dimMeta?.dimensionPath}
          {@const colStart = headerGroup.headers
            .slice(0, i)
            .reduce((sum, h) => sum + h.colSpan, 0)}
          {@const inHoverRange =
            hoveredColRange && isHeaderInHoveredRange(colStart, header.colSpan)}
          {@const isTheHoveredHeader =
            inHoverRange && isHoveredHeader(colStart, header.colSpan)}
          {@const isSelfSelected =
            isColDimHeader &&
            !!dimMeta.dimensionPath &&
            (clickSelection?.isColumnHeaderSelected(dimMeta.dimensionPath) ??
              false)}
          {@const inSelectedRange = isInSelectedColRange(
            colStart,
            header.colSpan,
            isSelfSelected,
          )}
          {@const inCellSelectedCol = isInCellSelectedColRange(
            colStart,
            header.colSpan,
          )}

          <th
            colSpan={header.colSpan}
            class:col-dim-hover-self={isTheHoveredHeader}
            class:col-dim-hover-child={inHoverRange && !isTheHoveredHeader}
            class:selected-col-header={isSelfSelected}
            class:in-selected-col-range={inSelectedRange}
            class:cell-selected-col-header={inCellSelectedCol}
            onmouseenter={() => {
              if (isColDimHeader) {
                hoveredColRange = {
                  start: colStart,
                  size: header.colSpan,
                };
              }
            }}
            onmouseleave={() => {
              hoveredColRange = null;
            }}
          >
            <button
              class="header-cell"
              class:cursor-pointer={header.column.getCanSort() ||
                (isColDimHeader && !!onColumnHeaderClick)}
              class:select-none={header.column.getCanSort()}
              class:flex-row-reverse={isMeasureColumn(header, i)}
              class:border-r={shouldShowHeaderRightBorder(header, i)}
              onclick={(e) => {
                if (isColDimHeader && onColumnHeaderClick) {
                  onColumnHeaderClick(dimMeta.dimensionPath ?? {});
                } else {
                  header.column.getToggleSortingHandler()?.(e);
                }
              }}
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
    <tr style:height="{before}px"></tr>
    {#each virtualRows as row (row.index)}
      {@const cells = rows[row.index].getVisibleCells()}
      {@const rowId = rows[row.index].id}
      {@const isTotalsRow = !!totalsRow && rowId === "0"}
      {@const isSelected = rowSelectionState?.isRowSelected(rowId) ?? false}
      {@const hasSelection = rowSelectionState?.hasActiveSelection ?? false}
      {@const isRowHeaderSelected =
        clickSelection?.isRowHeaderSelected(rowId) ?? false}
      {@const hasClickedCell =
        clickSelection?.hasSelectedCellInRow(rowId) ?? false}
      <tr
        class:show-more-row={isShowMoreRow(rows[row.index])}
        class:selected-row={isSelected && isRowHeaderSelected}
        class:dimmed-row={hasSelection && !isSelected && !hasClickedCell}
      >
        {#each cells as cell, i (cell.id)}
          {@const result =
            typeof cell.column.columnDef.cell === "function"
              ? cell.column.columnDef.cell(cell.getContext())
              : cell.column.columnDef.cell}
          {@const isActive = isCellActive(cell)}
          {@const isClicked = isCellClicked(cell)}
          {@const inHoveredCol =
            hoveredColRange && isHeaderInHoveredRange(i, 1)}
          {@const inSelectedCol = selectedColIndices.has(i)}
          {@const tooltipValue = cell.column.columnDef.meta?.tooltipFormatter
            ? cell.column.columnDef.meta.tooltipFormatter(cell.getValue())
            : cell.getValue()}
          <td
            class="ui-copy-number cell truncate group/cell"
            class:active-cell={isActive}
            class:selected-cell={isClicked}
            class:col-dim-hover-body={inHoveredCol}
            class:selected-col-body={inSelectedCol}
            class:cell-selected-row-header={i === 0 && hasClickedCell}
            class:interactive-cell={isTotalsRow
              ? canShowDataViewer
              : canShowDataViewer || enableClickToFilter}
            class:border-r={shouldShowRightBorder(i)}
            data-value={tooltipValue}
            data-rowid={cell.row.id}
            data-columnid={cell.column.id}
            data-rowheader={i === 0 || undefined}
            class:totals-column={i > 0 && i <= measureCount}
            onmouseover={() =>
              cellInspectorStore.updateValue(cell.getValue(), tooltipValue)}
            onfocus={() =>
              cellInspectorStore.updateValue(cell.getValue(), tooltipValue)}
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
    <tr style:height="{after}px"></tr>
  </tbody>
</table>

<style lang="postcss">
  * {
    @apply border-gray-200;
  }

  .resize-bar {
    @apply bg-primary-500 w-1 h-full;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal cursor-default;
    @apply bg-surface-background table-fixed;
  }

  /* Pin header */
  thead {
    @apply sticky top-0;
    @apply z-30 bg-surface-background;
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
    @apply relative;
  }

  thead tr:last-of-type th {
    @apply border-b;
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
    @apply px-2 size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    @apply text-fg-primary font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2 text-fg-primary;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  .with-row-dimension thead > tr:not(:last-of-type) > th:first-of-type {
    @apply border-b-0;
  }

  .with-row-dimension tr > th:first-of-type {
    @apply sticky left-0 z-20;
    @apply bg-surface-base;
  }

  .with-row-dimension tr > td:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-surface-base;
  }

  .with-row-dimension tr:hover > td:first-of-type {
    @apply bg-surface-hover;
  }

  .with-row-dimension.with-col-dimension tr > th:first-of-type {
    @apply bg-surface-background;
  }

  /* The totals row */
  .with-totals-row tbody > tr:nth-of-type(2) {
    @apply bg-surface-background sticky z-20;
    top: var(--total-header-height);
    height: calc(var(--row-height) + 2px);
  }

  /* The totals row header - only apply when there are actual measures and totals */
  .with-row-dimension.with-totals-row.with-measures
    tbody
    > tr:nth-of-type(2)
    > td:first-of-type {
    @apply font-semibold bg-surface-background;
  }

  .with-expandable-rows.with-totals-row
    tbody
    > tr:nth-of-type(2)
    > td:first-of-type {
    @apply pl-5;
  }

  tbody tr:hover,
  tbody tr:hover .cell {
    @apply bg-surface-hover;
  }

  tbody tr:hover .active-cell {
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

  .selected-cell.cell {
    @apply bg-primary-50 ring-1 ring-inset ring-primary-400;
  }
  .selected-cell.cell:hover {
    @apply bg-primary-100;
  }

  .selected-row .cell {
    @apply bg-primary-50;
  }
  .selected-row:hover .cell {
    @apply bg-primary-100;
  }
  .with-row-dimension .selected-row > td:first-of-type {
    @apply bg-primary-100;
  }
  .with-row-dimension .selected-row:hover > td:first-of-type {
    @apply bg-primary-100;
  }

  .dimmed-row .cell {
    @apply opacity-50;
  }

  /* Show more row styling */
  .show-more-row,
  .show-more-row .cell {
    @apply bg-surface-background;
  }

  .show-more-row:hover,
  .show-more-row:hover .cell {
    @apply bg-gray-100;
  }

  /* Column dimension header hover: the hovered header itself */
  .col-dim-hover-self .header-cell {
    @apply bg-primary-100;
  }

  /* Column dimension header hover: child headers below the hovered one */
  .col-dim-hover-child .header-cell,
  .col-dim-hover-body.cell {
    @apply bg-primary-50;
  }

  .selected-col-header .header-cell {
    @apply bg-primary-100;
  }

  .in-selected-col-range .header-cell,
  .selected-col-body.cell {
    @apply bg-primary-50;
  }

  /* Cross-highlights for cell click-to-filter selections */
  .cell-selected-col-header .header-cell {
    @apply bg-primary-50;
  }
  .cell-selected-col-header.col-dim-hover-self .header-cell {
    @apply bg-primary-100;
  }

  .cell-selected-row-header.cell {
    @apply bg-primary-50;
  }
  .with-row-dimension tr > td.cell-selected-row-header:first-of-type {
    @apply bg-primary-50;
  }
  .with-row-dimension tr:hover > td.cell-selected-row-header:first-of-type {
    @apply bg-primary-100;
  }
</style>
