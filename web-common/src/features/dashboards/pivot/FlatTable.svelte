<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type { MeasureColumnProps } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import {
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
    calculateColumnWidth,
    calculateMeasureWidth,
    clampColumnWidth,
    fitColumnWidthsToContainer,
    layoutColumnWidths,
    roleWidthBounds,
    type ColumnWidthRole,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import PivotExpandableCell from "@rilldata/web-common/features/dashboards/pivot/PivotExpandableCell.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { writable } from "svelte/store";
  import type {
    Cell,
    Column,
    HeaderGroup,
    Row,
  } from "tanstack-table-8-svelte-5";
  import { flexRender } from "tanstack-table-8-svelte-5";
  import { cellInspectorStore } from "../stores/cell-inspector-store";
  import {
    computeEffectiveDimIdx,
    flatCellState,
    flatRowState,
  } from "./pivot-cell-classes";
  import {
    type PivotClickSelectionState,
    dimKeyFromRow,
  } from "./pivot-click-selection";
  import type { PivotRowSelectionState } from "./pivot-row-selection";
  import type { CellFormatter } from "./pivot-conditional-formatting";
  import PivotHeaderLabel from "./PivotHeaderLabel.svelte";
  import type {
    PivotColumnAlign,
    PivotColumnStyles,
    PivotDataRow,
    PivotDataStoreConfig,
    PivotTotalsRowPosition,
  } from "./types";

  // State props
  export let assembled: boolean;
  export let measures: MeasureColumnProps;
  export let cellFormatters: Map<string, CellFormatter> = new Map();
  export let dataRows: PivotDataRow[];
  export let hasMeasureContextColumns: boolean;
  export let canShowDataViewer = false;
  export let enableClickToFilter = false;
  export let rowSelectionState: PivotRowSelectionState | undefined = undefined;
  export let clickSelection: PivotClickSelectionState | undefined = undefined;
  export let activeCell: { rowId: string; columnId: string } | null | undefined;
  export let config: PivotDataStoreConfig | undefined = undefined;
  export let fillWidth = false;
  export let containerWidth = 0;
  export let headerHeight = 30;
  // Per-column presentation overrides keyed by column id.
  export let columnStyles: PivotColumnStyles = {};
  export let fitToWidth = false;
  export let wrapText = false;
  export let wrapHeaders = false;
  export let onColumnResizeEnd:
    | ((columnId: string, width: number | null) => void)
    | undefined = undefined;

  // Table props
  export let headerGroups: HeaderGroup<PivotDataRow>[];
  export let rows: Row<PivotDataRow>[];
  export let virtualRows: { index: number }[];
  // The grand-totals row is not part of `rows`; it is a standalone tanstack row
  // pinned to the top of the body or rendered in a sticky <tfoot>.
  export let totalsRow: Row<PivotDataRow> | undefined;
  export let totalsRowPosition: PivotTotalsRowPosition = "top";
  export let before: number;
  export let after: number;
  export let totalRowSize: number;

  // Event handlers
  export let onCellClick: (e: MouseEvent) => void;
  export let onMouseMove: (e: MouseEvent) => void;
  export let onTableLeave: () => void;
  export let onCellCopy: (e: MouseEvent) => void;

  const columnLengths = writable(new Map<string, number>());
  // First estimate per column, used as the double-click reset width.
  const estimatedLengths = new Map<string, number>();

  let totalLength = 0;

  $: headers = headerGroups[0].headers;

  $: totalsRowAtBottom = totalsRowPosition === "bottom";

  $: timeDimension = config?.time?.timeDimension ?? "";
  $: measureIds = new Set(measures.map((m) => m.name));

  // Initialize column lengths if not already set: the configured width when
  // there is one, else an estimate from the data.
  $: headers.forEach((header) => {
    const columnId = header.column.id;

    if (!$columnLengths.has(columnId)) {
      const measure = getMeasureColumn(header.column);
      const estimatedWidth = measure
        ? calculateMeasureWidth(
            measure.name,
            measure.label,
            measure.formatter,
            totalsRow?.original,
            dataRows,
          )
        : calculateColumnWidth(
            columnId,
            String(header.column.columnDef.header),
            timeDimension,
            dataRows,
          );
      estimatedLengths.set(columnId, estimatedWidth);
      const configured = columnStyles[columnId]?.width;
      const initialWidth =
        configured === undefined
          ? estimatedWidth
          : clampColumnWidth(measure ? "measure" : "dimension", configured);
      columnLengths.update((lengths) => lengths.set(columnId, initialWidth));
    }
  });

  // Re-apply configured widths when they change (e.g. the YAML is edited while
  // the table is open) without touching columns the user resized. Compares
  // against the last applied values so unrelated spec emissions are no-ops,
  // and does not read the width store so it never re-runs per drag frame.
  let appliedConfiguredWidths: Record<string, number> = {};
  $: applyConfiguredWidths(columnStyles);
  function applyConfiguredWidths(styles: PivotColumnStyles) {
    const next: Record<string, number> = {};
    for (const [id, style] of Object.entries(styles)) {
      if (typeof style?.width === "number") next[id] = style.width;
    }
    const changed = Object.keys(next).filter(
      (id) => appliedConfiguredWidths[id] !== next[id],
    );
    const removed = Object.keys(appliedConfiguredWidths).filter(
      (id) => !(id in next),
    );
    appliedConfiguredWidths = next;
    if (!changed.length && !removed.length) return;
    columnLengths.update((lengths) => {
      for (const id of changed) {
        lengths.set(id, clampColumnWidth(roleOf(id), next[id]));
      }
      // Dropping the entry makes the seeding above estimate the width again.
      for (const id of removed) lengths.delete(id);
      return lengths;
    });
  }

  // Configured columns and columns the user resized in this session keep
  // their width: stretch and fit only move the other columns.
  let draggedIds = new Set<string>();
  $: pinnedIds = new Set<string>([
    ...Object.keys(columnStyles).filter(
      (id) => typeof columnStyles[id]?.width === "number",
    ),
    ...draggedIds,
  ]);

  $: baseColumnWidths = headers.map(
    (header) =>
      $columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH,
  );
  $: displayColumnWidths = layoutColumnWidths(
    headers.map((header, i) => ({
      width: baseColumnWidths[i],
      role: measureIds.has(header.column.id) ? "measure" : "dimension",
      pinned: pinnedIds.has(header.column.id),
    })),
    containerWidth,
    { fill: fillWidth, fit: fitToWidth },
  );
  $: totalLength = displayColumnWidths.reduce((acc, width) => {
    return acc + width;
  }, 0);

  // Alignment and wrapping per column id, derived here (not in template
  // functions) so the template re-renders when the styles change.
  $: columnPresentation = new Map<
    string,
    { align: PivotColumnAlign; wrap: boolean }
  >(
    headers.map((header) => {
      const id = header.column.id;
      const isMeasure = measureIds.has(id);
      const style = columnStyles[id];
      return [
        id,
        {
          align: style?.align ?? (isMeasure ? "right" : "left"),
          // Measure cells never wrap.
          wrap: !isMeasure && (style?.wrap ?? wrapText),
        },
      ];
    }),
  );

  function getMeasureColumn(headerColumn: Column<PivotDataRow>) {
    const columnId = headerColumn.id;
    return measures.find((m) => m.name === columnId);
  }

  function roleOf(columnId: string): ColumnWidthRole {
    return measureIds.has(columnId) ? "measure" : "dimension";
  }

  function markDragged(columnId: string) {
    if (draggedIds.has(columnId)) return;
    draggedIds = new Set(draggedIds).add(columnId);
  }

  function unmarkDragged(columnId: string) {
    if (!draggedIds.has(columnId)) return;
    const next = new Set(draggedIds);
    next.delete(columnId);
    draggedIds = next;
  }

  /**
   * One-shot resize of every column so the table fits the given width.
   * Writes the fitted widths into the column length store, so subsequent
   * manual resizes start from the fitted widths and nothing re-fits.
   */
  export function fitColumnsToWidth(availableWidth: number) {
    const fitted = fitColumnWidthsToContainer(
      headers.map((header) => ({
        width:
          $columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH,
        ...roleWidthBounds(roleOf(header.column.id)),
      })),
      availableWidth,
    );
    columnLengths.update((lengths) => {
      headers.forEach((header, i) => lengths.set(header.column.id, fitted[i]));
      return lengths;
    });
  }

  // Resolve conditional-formatting styling for a measure cell. Returns null for
  // cells that should not be formatted (no config, the totals row, non-numeric
  // values, or no matching threshold rule).
  function getCellFormatting(
    cell: Cell<PivotDataRow, unknown>,
    isTotalsRow: boolean,
  ): { background: string; color: string } | null {
    if (isTotalsRow) return null;
    const meta = cell.column.columnDef.meta;
    if (!meta?.conditionalFormat || meta.isRowTotal || !meta.measureName) {
      return null;
    }
    const formatter = cellFormatters.get(meta.measureName);
    if (!formatter) return null;
    const value = cell.getValue();
    if (typeof value !== "number" || !Number.isFinite(value)) return null;
    return formatter(value);
  }

  function isCellActive(rowId: string, columnId: string) {
    return rowId === activeCell?.rowId && columnId === activeCell?.columnId;
  }

  $: lastDimIdx = (config?.rowDimensionNames.length ?? 0) - 1;

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
  style:height="{totalRowSize + headerHeight + headerGroups.length}px"
>
  {#each headers as header, i (header.id)}
    {@const columnId = header.column.id}
    {@const baseLength =
      $columnLengths.get(columnId) ?? WIDTHS.INIT_MEASURE_WIDTH}
    {@const length = displayColumnWidths[i] ?? baseLength}
    {@const last = i === headers.length - 1}
    {@const bounds = roleWidthBounds(
      measureIds.has(columnId) ? "measure" : "dimension",
    )}
    <div style:width="{length}px" class="h-full relative">
      <Resizer
        side="right"
        direction="EW"
        min={bounds.min}
        max={bounds.max}
        basis={columnStyles[columnId]?.width ??
          estimatedLengths.get(columnId) ??
          WIDTHS.INIT_MEASURE_WIDTH}
        dimension={baseLength}
        justify={last ? "end" : "center"}
        hang={!last}
        onUpdate={(d: number) => {
          // Pinned once the pointer has actually moved, not on mousedown.
          markDragged(columnId);
          columnLengths.update((lengths) => lengths.set(columnId, d));
        }}
        onMouseUp={(d: number, moved: boolean) => {
          // A click without movement must neither persist nor pin the column.
          if (moved) onColumnResizeEnd?.(columnId, d);
        }}
        onReset={() => {
          unmarkDragged(columnId);
          onColumnResizeEnd?.(columnId, null);
        }}
      >
        <div class="resize-bar"></div>
      </Resizer>
    </div>
  {/each}
</div>

<table
  role="presentation"
  style:width="{totalLength}px"
  onclick={modified({ shift: onCellCopy, click: onCellClick })}
  onmousemove={onMouseMove}
  onmouseleave={onTableLeave}
>
  <colgroup>
    {#each headers as header, i (header.id)}
      {@const baseLength =
        $columnLengths.get(header.column.id) ?? WIDTHS.INIT_MEASURE_WIDTH}
      {@const length = displayColumnWidths[i] ?? baseLength}
      <col style:width="{length}px" style:max-width="{length}px" />
    {/each}
  </colgroup>

  <thead>
    {#each headerGroups as headerGroup (headerGroup.id)}
      <tr>
        {#each headerGroup.headers as header (header.id)}
          {@const sortDirection = header.column.getIsSorted()}
          {@const icon = header.column.columnDef.meta?.icon}
          {@const presentation = columnPresentation.get(header.column.id)}
          <th>
            <button
              class="header-cell"
              class:cursor-pointer={header.column.getCanSort()}
              class:select-none={header.column.getCanSort()}
              class:flex-row-reverse={presentation?.align === "right"}
              class:justify-center={presentation?.align === "center"}
              class:border-r={hasBorderRight(header.column.id)}
              onclick={header.column.getToggleSortingHandler()}
            >
              {#if !header.isPlaceholder}
                {#if icon}
                  <svelte:component this={icon} />
                {:else}
                  <PivotHeaderLabel
                    label={String(header.column.columnDef.header)}
                    description={header.column.columnDef.meta?.description}
                    wrap={wrapHeaders}
                  />
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
    {#if totalsRow && !totalsRowAtBottom}
      {@render pivotRow(totalsRow, true)}
    {/if}
    {#each virtualRows as virtualRow (virtualRow.index)}
      {@render pivotRow(rows[virtualRow.index], false)}
    {/each}
    <tr style:height="{after}px"></tr>
  </tbody>
  {#if totalsRow && totalsRowAtBottom}
    <tfoot>
      {@render pivotRow(totalsRow, true)}
    </tfoot>
  {/if}
</table>

{#snippet pivotRow(row: Row<PivotDataRow>, isTotalsRow: boolean)}
  {@const cells = row.getVisibleCells()}
  {@const rowData = row.original}
  {@const dk = dimKeyFromRow(rowData, config?.rowDimensionNames ?? [])}
  {@const isSelected = rowSelectionState?.isRowSelected(rowData) ?? false}
  {@const hasClickedCell = clickSelection?.hasSelectedCellInRow(dk) ?? false}
  {@const effectiveDimIdx = computeEffectiveDimIdx(
    hasClickedCell,
    clickSelection?.getClickedDimensionIndex(dk) ?? -1,
    lastDimIdx,
    isSelected,
    rowSelectionState?.maxFilteredDimensionIndex ?? -1,
  )}
  {@const rs = flatRowState({
    isSelected,
    hasSelection: rowSelectionState?.hasActiveSelection ?? false,
    hasClickedCell,
    effectiveDimIdx,
  })}
  <tr
    class:totals-row={isTotalsRow}
    class:selected-row={rs.selectedRow}
    class:dimmed-row={rs.dimmedRow}
  >
    {#each cells as cell (cell.id)}
      {@const result =
        typeof cell.column.columnDef.cell === "function"
          ? cell.column.columnDef.cell(cell.getContext())
          : cell.column.columnDef.cell}
      {@const cs = flatCellState({
        isActive: isCellActive(cell.row.id, cell.column.id),
        isClicked: clickSelection?.isCellSelected(dk, cell.column.id) ?? false,
        colDimIdx: config?.rowDimensionNames.indexOf(cell.column.id) ?? -1,
        effectiveDimIdx,
        lastDimIdx,
        isTotalsRow,
        canShowDataViewer,
        enableClickToFilter,
        hasValue: cell.getValue() !== undefined,
      })}
      {@const tooltipValue = cell.column.columnDef.meta?.tooltipFormatter
        ? cell.column.columnDef.meta.tooltipFormatter(cell.getValue())
        : cell.getValue()}
      {@const cellFmt = getCellFormatting(cell, isTotalsRow)}
      {@const presentation = columnPresentation.get(cell.column.id)}
      {@const wrapCell = presentation?.wrap ?? false}
      <td
        class="ui-copy-number cell"
        class:truncate={!wrapCell}
        class:wrap-cell={wrapCell}
        class:has-conditional-format={cellFmt !== null}
        style:--cf-bg={cellFmt?.background ?? null}
        style:--cf-color={cellFmt?.color ?? null}
        style:text-align={presentation?.align ?? null}
        class:active-cell={cs.activeCell}
        class:selected-cell={cs.selectedCell}
        class:selected-context-cell={cs.selectedContextCell}
        class:muted-cell={cs.mutedCell}
        class:interactive-cell={cs.interactiveCell}
        class:border-r={hasBorderRight(cell.column.id)}
        class:total-label={cell.getValue() === "Total"}
        data-value={tooltipValue}
        data-rowid={cell.row.id}
        data-columnid={cell.column.id}
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
            {...wrapCell && result.component === PivotExpandableCell
              ? { wrap: true }
              : {}}
          />
        {:else if typeof result === "string" || typeof result === "number"}
          {#if wrapCell}
            <span class="wrap-text">{result}</span>
          {:else}
            {result}
          {/if}
        {:else}
          <svelte:component
            this={flexRender(cell.column.columnDef.cell, cell.getContext())}
          />
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}

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

  tbody .cell,
  tfoot .cell {
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
    @apply px-2 bg-surface-background size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    @apply text-fg-primary font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2 text-fg-primary;
  }

  /* Wrapped dimension cells: the row keeps its fixed height and the text is
     clamped to --wrap-lines lines */
  td.wrap-cell {
    @apply whitespace-normal;
  }

  .wrap-text {
    @apply whitespace-normal overflow-hidden;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: var(--wrap-lines, 2);
    overflow-wrap: anywhere;
    line-height: 1rem;
  }

  /* Conditional formatting (heatmap / data bar). Placed before the
     selection/hover rules below and kept at equal specificity (0,2,0) so
     those win by source order, keeping selected and hovered cells legible. */
  .cell.has-conditional-format {
    background: var(--cf-bg);
    color: var(--cf-color);
  }

  /* When a hover/selection state replaces the conditional background with a
     light surface color, the formatter's text color (which may be white, chosen
     for a dark heatmap fill) becomes illegible. Revert to the default
     foreground so the value stays readable. */
  tr:hover td.cell.has-conditional-format,
  td.cell.has-conditional-format.active-cell,
  td.cell.has-conditional-format.selected-cell,
  td.cell.has-conditional-format.selected-context-cell,
  td.cell.has-conditional-format.muted-cell,
  .selected-row td.cell.has-conditional-format {
    @apply text-fg-primary;
  }

  tr > td {
    @apply font-normal;
  }

  /* The totals row: pinned under the header, or above the bottom edge when
     rendered in the tfoot */
  tbody > tr.totals-row {
    @apply bg-surface-background sticky z-20;
    top: var(--total-header-height);
  }
  tfoot > tr.totals-row {
    @apply bg-surface-background sticky bottom-0 z-20;
  }
  tfoot > tr.totals-row > td {
    @apply border-t;
  }

  /* The totals row label - make it bold for flat tables */
  .total-label {
    @apply font-semibold;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-surface-hover;
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

  td.selected-cell.cell {
    @apply bg-primary-50 relative z-[1];
    box-shadow: 0 0 0 1px theme(colors.primary.400);
  }
  /* The totals row is z-20 and covers the outset top shadow; use an inset top border instead */
  tbody > tr.totals-row + tr > td.selected-cell.cell {
    box-shadow:
      0 0 0 1px theme(colors.primary.400),
      inset 0 1px 0 0 theme(colors.primary.400);
  }
  td.selected-cell.cell:hover {
    @apply bg-primary-100;
  }

  .selected-row .cell {
    @apply bg-primary-50;
  }
  .selected-row:hover .cell {
    @apply bg-primary-100;
  }

  .dimmed-row .cell {
    @apply opacity-50;
  }

  /* Dimension cells to the left of the clicked cell: same primary background, no ring */
  .selected-context-cell.cell {
    @apply bg-primary-50;
  }
  .selected-context-cell.cell:hover {
    @apply bg-primary-100;
  }

  /* Dimension cells to the right of the clicked cell: muted background */
  .muted-cell.cell {
    @apply bg-surface-muted;
  }
</style>
