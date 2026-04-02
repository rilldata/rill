/**
 * Pure functions for computing column/row selection indices from
 * PivotClickSelectionState and TanStack header groups. Extracted
 * from NestedTable.svelte's reactive $: blocks to enable testing
 * and reuse.
 */

import type { HeaderGroup, Row } from "tanstack-table-8-svelte-5";
import type { PivotClickSelectionState } from "./pivot-click-selection";
import { dimKeyFromRow, nestedDimKeyFromRow } from "./pivot-click-selection";
import type { PivotDataRow } from "./types";

/**
 * Compute the set of leaf column indices that fall within any
 * selected column header's span. Iterates all header groups to
 * find headers whose dimensionPath matches the selection.
 */
export function computeSelectedColIndices(
  clickSelection: PivotClickSelectionState | undefined,
  headerGroups: HeaderGroup<PivotDataRow>[],
): Set<number> {
  if (!clickSelection?.hasAnySelection) return new Set();
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
}

/**
 * Compute the set of leaf column indices that have at least one
 * cell selected via click-to-filter. Uses the last header group
 * (leaf columns) to map column IDs to indices.
 */
export function computeCellSelectedColIndices(
  clickSelection: PivotClickSelectionState | undefined,
  headerGroups: HeaderGroup<PivotDataRow>[],
): Set<number> {
  if (!clickSelection?.selectedCellColumnIds?.size) return new Set();
  const leafGroup = headerGroups[headerGroups.length - 1];
  if (!leafGroup) return new Set();
  const indices = new Set<number>();
  let colIdx = 0;
  for (const header of leafGroup.headers) {
    if (clickSelection.selectedCellColumnIds.has(header.column.id)) {
      indices.add(colIdx);
    }
    colIdx += header.colSpan;
  }
  return indices;
}

/**
 * Compute TanStack row IDs that are ancestors of any selected
 * child row (row header click or cell click). Used to highlight
 * parent row headers in nested tables.
 */
export function computeAncestorRowIds(
  clickSelection: PivotClickSelectionState | undefined,
  rows: Row<PivotDataRow>[],
  rowDimensionNames: string[],
): Set<string> {
  if (!clickSelection?.hasAnySelection) return new Set();
  const ancestorIds = new Set<string>();
  for (const row of rows) {
    const dk =
      row.depth > 0
        ? nestedDimKeyFromRow(row, rowDimensionNames)
        : dimKeyFromRow(row.original, rowDimensionNames);
    if (
      clickSelection.isRowHeaderSelected(dk) ||
      clickSelection.hasSelectedCellInRow(dk)
    ) {
      let id = row.id;
      while (id.includes(".")) {
        id = id.substring(0, id.lastIndexOf("."));
        ancestorIds.add(id);
      }
    }
  }
  return ancestorIds;
}

// ---- Column header hover/selection range helpers ----

export interface HoveredColRange {
  start: number;
  size: number;
}

/** Check if a header (by its leaf column range) falls within the hovered range */
export function isHeaderInHoveredRange(
  headerStart: number,
  headerSize: number,
  hoveredColRange: HoveredColRange | null,
): boolean {
  if (!hoveredColRange) return false;
  const hovEnd = hoveredColRange.start + hoveredColRange.size;
  return (
    headerStart >= hoveredColRange.start && headerStart + headerSize <= hovEnd
  );
}

/** Check if a header IS the exact hovered header */
export function isHoveredHeader(
  colStart: number,
  colSpan: number,
  hoveredColRange: HoveredColRange | null,
): boolean {
  if (!hoveredColRange) return false;
  return colStart === hoveredColRange.start && colSpan === hoveredColRange.size;
}

/**
 * Check if a header should be highlighted as falling within a
 * selected column range (child of a clicked column header).
 */
export function isInSelectedColRange(
  colStart: number,
  colSpan: number,
  isSelfSelected: boolean,
  selectedColIndices: Set<number>,
): boolean {
  if (selectedColIndices.size === 0 || colSpan === 0 || isSelfSelected) {
    return false;
  }
  for (let i = colStart; i < colStart + colSpan; i++) {
    if (!selectedColIndices.has(i)) return false;
  }
  return true;
}

/** Check if a header (by its leaf column range) contains any cell-selected columns */
export function isInCellSelectedColRange(
  colStart: number,
  colSpan: number,
  cellSelectedColIndices: Set<number>,
): boolean {
  if (cellSelectedColIndices.size === 0) return false;
  for (let i = colStart; i < colStart + colSpan; i++) {
    if (cellSelectedColIndices.has(i)) return true;
  }
  return false;
}
