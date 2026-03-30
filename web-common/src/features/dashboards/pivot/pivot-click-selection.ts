/**
 * Tracks which pivot elements (row headers, data cells, column headers)
 * were selected via click-to-filter.
 *
 * Selections are keyed by dimension values (dimKey), not positional row
 * indices, so they remain stable across sorting and data refreshes.
 */

import type { PivotDataRow } from "./types";

/**
 * Produces a stable string key from a row's dimension values.
 * Uses NUL as separator since dimension values won't contain it.
 */
export function dimKeyFromRow(
  rowData: PivotDataRow,
  rowDimensionNames: string[],
): string {
  return rowDimensionNames.map((d) => String(rowData[d] ?? "")).join("\0");
}

export function cellKey(dimKey: string, columnId: string) {
  return `${dimKey}\t${columnId}`;
}

export function columnHeaderKey(dimensionPath: Record<string, string>): string {
  return JSON.stringify(
    Object.entries(dimensionPath).sort(([a], [b]) => a.localeCompare(b)),
  );
}

// ---- Selection entry (stored per cell / row-header click) ----

export interface SelectionEntry {
  dimKey: string;
  /** Row dimension name→value pairs captured at click time */
  dimValues: Record<string, string>;
  columnId: string;
  /** For flat-table dimension cell clicks: the index into rowDimensionNames */
  dimClickIndex?: number;
}

// ---- Selection state ----

export interface PivotClickSelectionState {
  /** dimKey → entry for row-header clicks */
  rowHeaderSelections: Map<string, SelectionEntry>;
  /** "dimKey\tcolumnId" → entry for data-cell clicks */
  cellSelections: Map<string, SelectionEntry>;
  /** Serialised dimension-path keys for column-header clicks */
  columnHeaderSelections: Set<string>;
  /** Whether any selection exists at all */
  hasAnySelection: boolean;
  /** Check if a specific row was selected via row-header click */
  isRowHeaderSelected: (dimKey: string) => boolean;
  /** Check if a specific cell was selected via data-cell click */
  isCellSelected: (dimKey: string, columnId: string) => boolean;
  /** Check if any data cell in this row was selected via click */
  hasSelectedCellInRow: (dimKey: string) => boolean;
  /** Check if a column header was selected via click */
  isColumnHeaderSelected: (dimensionPath: Record<string, string>) => boolean;
  /** Column IDs that have at least one selected cell (for highlighting column headers) */
  selectedCellColumnIds: Set<string>;
  /**
   * Returns the dimension column index that was clicked in this row
   * (i.e. the index into rowDimensionNames), or -1 if no dimension cell
   * was clicked (measure click, row-header click, or no selection).
   */
  getClickedDimensionIndex: (dimKey: string) => number;
}

export function createEmptyClickSelectionState(): PivotClickSelectionState {
  return {
    rowHeaderSelections: new Map(),
    cellSelections: new Map(),
    columnHeaderSelections: new Set(),
    hasAnySelection: false,
    isRowHeaderSelected: () => false,
    isCellSelected: () => false,
    hasSelectedCellInRow: () => false,
    isColumnHeaderSelected: () => false,
    selectedCellColumnIds: new Set(),
    getClickedDimensionIndex: () => -1,
  };
}

export function buildClickSelection(
  rowHeaders: Map<string, SelectionEntry>,
  cells: Map<string, SelectionEntry>,
  colHeaders: Set<string>,
): PivotClickSelectionState {
  const hasAny = rowHeaders.size > 0 || cells.size > 0 || colHeaders.size > 0;

  // Build sets of dimKeys and columnIds that have at least one selected cell
  const rowsWithSelectedCells = new Set<string>();
  const columnsWithSelectedCells = new Set<string>();
  for (const entry of cells.values()) {
    rowsWithSelectedCells.add(entry.dimKey);
    columnsWithSelectedCells.add(entry.columnId);
  }

  // Build a map of dimKey → dimClickIndex for quick lookup
  const dimClickIndexByKey = new Map<string, number>();
  for (const entry of cells.values()) {
    if (entry.dimClickIndex !== undefined && entry.dimClickIndex >= 0) {
      dimClickIndexByKey.set(entry.dimKey, entry.dimClickIndex);
    }
  }

  return {
    rowHeaderSelections: rowHeaders,
    cellSelections: cells,
    columnHeaderSelections: colHeaders,
    hasAnySelection: hasAny,
    isRowHeaderSelected: (dk) => rowHeaders.has(dk),
    isCellSelected: (dk, cid) => cells.has(cellKey(dk, cid)),
    hasSelectedCellInRow: (dk) => rowsWithSelectedCells.has(dk),
    isColumnHeaderSelected: (path) => colHeaders.has(columnHeaderKey(path)),
    selectedCellColumnIds: columnsWithSelectedCells,
    getClickedDimensionIndex: (dk) => dimClickIndexByKey.get(dk) ?? -1,
  };
}
