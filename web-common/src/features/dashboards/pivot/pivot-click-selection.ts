/**
 * Tracks which pivot elements (row headers, data cells, column headers)
 * were selected via click-to-filter.
 *
 * Selections are keyed by dimension values (dimKey), not positional row
 * indices, so they remain stable across sorting and data refreshes.
 */

import type { Row } from "tanstack-table-8-svelte-5";
import type { PivotDataRow } from "./types";

// Distinct sentinel for null dimension values so a null at depth N does not
// collide with "no value at depth N" (e.g. a depth-0 row whose deeper
// dimensions are simply absent from rowData).
const NULL_KEY_SENTINEL = "<NULL>";

function encodeKeyValue(value: unknown): string {
  return value === null ? NULL_KEY_SENTINEL : String(value);
}

/**
 * Produces a stable string key from a row's dimension values.
 * Uses NUL as separator since dimension values won't contain it.
 *
 * Truncates to the deepest dimension that is actually present in rowData,
 * so a depth-0 row in a nested table (which only stores rowDimensionNames[0])
 * does not collide with a deeper row whose later dimensions happen to be null.
 *
 * WARNING: In nested tables, rows only store their own value under
 * rowDimensionNames[0], so this function will NOT include parent
 * dimension values. Use dimKeyFromDimValues instead for nested tables.
 */
export function dimKeyFromRow(
  rowData: PivotDataRow,
  rowDimensionNames: string[],
): string {
  let lastIdx = rowDimensionNames.length - 1;
  while (lastIdx >= 0 && !(rowDimensionNames[lastIdx] in rowData)) {
    lastIdx--;
  }
  return rowDimensionNames
    .slice(0, lastIdx + 1)
    .map((d) => encodeKeyValue(rowData[d]))
    .join("\0");
}

/**
 * Produces a stable string key from resolved dimension name→value pairs.
 * For nested tables, callers should resolve all ancestor values first
 * (e.g. via getDimensionValuesForRow) so the key is unique across
 * different parents. Truncates to the deepest dimension that is actually
 * present in dimValues so depth-N keys do not collide with deeper rows
 * whose later dimensions are null.
 */
export function dimKeyFromDimValues(
  dimValues: Record<string, string | null>,
  rowDimensionNames: string[],
): string {
  let lastIdx = rowDimensionNames.length - 1;
  while (lastIdx >= 0 && !(rowDimensionNames[lastIdx] in dimValues)) {
    lastIdx--;
  }
  return rowDimensionNames
    .slice(0, lastIdx + 1)
    .map((d) => encodeKeyValue(dimValues[d]))
    .join("\0");
}

/**
 * Produces a stable dimKey for a TanStack Row in a nested table by
 * walking the parent chain. Each row at depth N stores its value under
 * rowDimensionNames[0]; the actual dimension is rowDimensionNames[depth].
 * Returns a NUL-separated key with one component per chain entry, so a
 * depth-0 row produces a 1-component key and a depth-N row produces an
 * (N+1)-component key.
 */
export function nestedDimKeyFromRow(
  row: Row<PivotDataRow>,
  rowDimensionNames: string[],
): string {
  const firstDim = rowDimensionNames[0];
  const chain = [...row.getParentRows(), row];
  return chain.map((r) => encodeKeyValue(r.original[firstDim])).join("\0");
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
  /** Row dimension name→value pairs captured at click time; null for null dimension values */
  dimValues: Record<string, string | null>;
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
  /** Whether both row-header and column-header selections exist (cross-selection) */
  hasCrossSelection: boolean;
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
  /** Check if a column header is an ancestor of any selected column header */
  isAncestorOfSelectedColumnHeader: (
    dimensionPath: Record<string, string>,
  ) => boolean;
}

export function createEmptyClickSelectionState(): PivotClickSelectionState {
  return {
    rowHeaderSelections: new Map(),
    cellSelections: new Map(),
    columnHeaderSelections: new Set(),
    hasAnySelection: false,
    hasCrossSelection: false,
    isRowHeaderSelected: () => false,
    isCellSelected: () => false,
    hasSelectedCellInRow: () => false,
    isColumnHeaderSelected: () => false,
    selectedCellColumnIds: new Set(),
    getClickedDimensionIndex: () => -1,
    isAncestorOfSelectedColumnHeader: () => false,
  };
}

export function buildClickSelection(
  rowHeaders: Map<string, SelectionEntry>,
  cells: Map<string, SelectionEntry>,
  colHeaders: Set<string>,
): PivotClickSelectionState {
  const hasAny = rowHeaders.size > 0 || cells.size > 0 || colHeaders.size > 0;
  const hasCrossSelection = rowHeaders.size > 0 && colHeaders.size > 0;

  // Build sets of dimKeys and columnIds that have at least one selected cell
  const rowsWithSelectedCells = new Set<string>();
  const columnsWithSelectedCells = new Set<string>();
  for (const entry of cells.values()) {
    rowsWithSelectedCells.add(entry.dimKey);
    columnsWithSelectedCells.add(entry.columnId);
  }

  // Pre-parse selected column header paths for ancestor checks
  const parsedColHeaders: Record<string, string>[] = [];
  for (const key of colHeaders) {
    const entries: [string, string][] = JSON.parse(key);
    parsedColHeaders.push(Object.fromEntries(entries));
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
    hasCrossSelection,
    isRowHeaderSelected: (dk) => rowHeaders.has(dk),
    isCellSelected: (dk, cid) => cells.has(cellKey(dk, cid)),
    hasSelectedCellInRow: (dk) => rowsWithSelectedCells.has(dk),
    isColumnHeaderSelected: (path) => colHeaders.has(columnHeaderKey(path)),
    selectedCellColumnIds: columnsWithSelectedCells,
    getClickedDimensionIndex: (dk) => dimClickIndexByKey.get(dk) ?? -1,
    isAncestorOfSelectedColumnHeader: (path) => {
      if (parsedColHeaders.length === 0) return false;
      const pathEntries = Object.entries(path);
      const pathSize = pathEntries.length;
      return parsedColHeaders.some((selectedPath) => {
        // Must be a strict superset (selected has more entries)
        if (Object.keys(selectedPath).length <= pathSize) return false;
        // Every entry in this header's path must exist in the selected path
        return pathEntries.every(([k, v]) => selectedPath[k] === v);
      });
    },
  };
}
