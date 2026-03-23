/**
 * Tracks which pivot elements (row headers, data cells, column headers)
 * were selected via click-to-filter. Pure data types and key builders
 * with no dependencies on other pivot modules.
 */

export interface PivotClickSelectionState {
  /** rowIds selected via row-header clicks */
  rowHeaderSelections: Set<string>;
  /** "rowId:columnId" keys selected via data-cell clicks */
  cellSelections: Set<string>;
  /** "dimensionName:dimensionValue" keys selected via column-header clicks */
  columnHeaderSelections: Set<string>;
  /** Whether any selection exists at all */
  hasAnySelection: boolean;
  /** Check if a specific row was selected via row-header click */
  isRowHeaderSelected: (rowId: string) => boolean;
  /** Check if a specific cell was selected via data-cell click */
  isCellSelected: (rowId: string, columnId: string) => boolean;
  /** Check if any data cell in this row was selected via click */
  hasSelectedCellInRow: (rowId: string) => boolean;
  /** Check if a column header was selected via click */
  isColumnHeaderSelected: (dimensionPath: Record<string, string>) => boolean;
  /** Column IDs that have at least one selected cell (for highlighting column headers) */
  selectedCellColumnIds: Set<string>;
  /**
   * Returns the dimension column index that was clicked in this row
   * (i.e. the index into rowDimensionNames), or -1 if no dimension cell
   * was clicked (measure click, row-header click, or no selection).
   */
  getClickedDimensionIndex: (rowId: string) => number;
}

export function createEmptyClickSelectionState(): PivotClickSelectionState {
  return {
    rowHeaderSelections: new Set(),
    cellSelections: new Set(),
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

export function cellKey(rowId: string, columnId: string) {
  return `${rowId}:${columnId}`;
}

export function columnHeaderKey(dimensionPath: Record<string, string>): string {
  return JSON.stringify(
    Object.entries(dimensionPath).sort(([a], [b]) => a.localeCompare(b)),
  );
}

export function buildClickSelection(
  rowHeaders: Set<string>,
  cells: Set<string>,
  colHeaders: Set<string>,
  /** Maps rowId → dimension column index for dimension-cell clicks in flat tables */
  rowDimClickIndex: Map<string, number> = new Map(),
): PivotClickSelectionState {
  const hasAny = rowHeaders.size > 0 || cells.size > 0 || colHeaders.size > 0;

  // Build sets of rowIds and columnIds that have at least one selected cell
  const rowsWithSelectedCells = new Set<string>();
  const columnsWithSelectedCells = new Set<string>();
  for (const key of cells) {
    const sep = key.indexOf(":");
    if (sep !== -1) {
      rowsWithSelectedCells.add(key.slice(0, sep));
      columnsWithSelectedCells.add(key.slice(sep + 1));
    }
  }

  return {
    rowHeaderSelections: rowHeaders,
    cellSelections: cells,
    columnHeaderSelections: colHeaders,
    hasAnySelection: hasAny,
    isRowHeaderSelected: (rid) => rowHeaders.has(rid),
    isCellSelected: (rid, cid) => cells.has(cellKey(rid, cid)),
    hasSelectedCellInRow: (rid) => rowsWithSelectedCells.has(rid),
    isColumnHeaderSelected: (path) => colHeaders.has(columnHeaderKey(path)),
    selectedCellColumnIds: columnsWithSelectedCells,
    getClickedDimensionIndex: (rid) => rowDimClickIndex.get(rid) ?? -1,
  };
}
