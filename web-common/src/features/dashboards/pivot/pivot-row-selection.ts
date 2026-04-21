import { getValuesForExpandedKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-expansion";
import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { getValuesInExpression } from "../stores/filter-utils";
import {
  buildPivotFilter,
  getValuesForFlatTable,
  isTimeDimension,
} from "./pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
  PivotFilter,
} from "./types";

export interface PivotRowSelectionState {
  /** True if at least one row dimension has active filter selections */
  hasActiveSelection: boolean;
  /**
   * Check if a row is selected based on current filters.
   * For nested tables, pass the row's depth and parentRows so the value
   * under rowDimensions[0] is checked against the correct dimension filter
   * and all ancestor dimensions are also verified.
   */
  isRowSelected: (
    rowData: PivotDataRow,
    depth?: number,
    parentRows?: PivotDataRow[],
  ) => boolean;
  /** Highest row dimension index with an active selection filter, or -1 */
  maxFilteredDimensionIndex: number;
}

/**
 * Returns the raw dimension values for a row, including time dimensions.
 * Shared by getDimensionValuesForRow (which filters out time) and
 * getFiltersForRowHeader (which passes entries to buildPivotFilter).
 */
function getRawRowValues(
  config: PivotDataStoreConfig,
  rowId: string,
  tableData: PivotDataRow[],
): string[] {
  const { rowDimensionNames, measureNames, isFlat } = config;
  return isFlat
    ? getValuesForFlatTable(
        tableData,
        rowDimensionNames,
        rowId,
        measureNames.length > 0,
      )
    : getValuesForExpandedKey(
        tableData,
        rowDimensionNames,
        rowId,
        measureNames.length > 0,
      );
}

/**
 * Extracts dimension name/value pairs for a given row without building
 * full V1Expression objects. Used for determining row selection state
 * and for click-to-filter. Filters out time dimensions.
 */
export function getDimensionValuesForRow(
  config: PivotDataStoreConfig,
  rowId: string,
  tableData: PivotDataRow[],
): Array<{ dimensionName: string; value: string }> {
  const values = getRawRowValues(config, rowId, tableData);

  return values
    .map((value, index) => ({
      dimensionName: config.rowDimensionNames[index],
      value,
    }))
    .filter(
      ({ dimensionName }) =>
        !isTimeDimension(dimensionName, config.time.timeDimension),
    );
}

/**
 * Like getDimensionValuesForRow but reads values directly from a PivotDataRow
 * instead of using positional rowId indexing. Stable across sorting.
 */
export function getDimensionValuesFromRowData(
  config: PivotDataStoreConfig,
  rowData: PivotDataRow,
): Array<{ dimensionName: string; value: string }> {
  return config.rowDimensionNames
    .filter((dim) => !isTimeDimension(dim, config.time.timeDimension))
    .map((dim) => ({
      dimensionName: dim,
      value: String(rowData[dim] ?? ""),
    }))
    .filter(({ value }) => value !== "");
}

/**
 * Like getFiltersForRowHeader but reads values directly from a PivotDataRow.
 * Stable across sorting.
 */
export function getFiltersForRowData(
  config: PivotDataStoreConfig,
  rowData: PivotDataRow,
): PivotFilter {
  const dimEntries = config.rowDimensionNames
    .filter((dim) => rowData[dim] !== undefined)
    .map((dim) => ({
      name: dim,
      value: rowData[dim] === null ? null : String(rowData[dim]),
    }));
  return buildPivotFilter(config, dimEntries);
}

/**
 * Extracts dimension -> Set<values> from a V1Expression (whereFilter).
 * Only considers IN expressions for selection matching; other filter types
 * (LIKE, NIN, threshold) are ignored for highlighting purposes.
 * Filters to only include dimensions present in rowDimensionNames.
 */
export function extractSelectionDimensionFilters(
  whereFilter: V1Expression | undefined,
  rowDimensionNames: string[],
): Map<string, Set<string>> {
  const result = new Map<string, Set<string>>();
  if (!whereFilter?.cond?.exprs) return result;

  for (const expr of whereFilter.cond.exprs) {
    if (expr.cond?.op !== V1Operation.OPERATION_IN) continue;

    const ident = expr.cond.exprs?.[0]?.ident;
    if (!ident || !rowDimensionNames.includes(ident)) continue;

    const values = getValuesInExpression(expr);
    if (values.length > 0) {
      result.set(ident, new Set(values as string[]));
    }
  }

  return result;
}

/**
 * Computes selection state for pivot rows given current dimension filters.
 *
 * A row is "selected" if every one of its row dimensions that has an active
 * filter contains the row's value for that dimension. Dimensions without
 * active filters are treated as matching (they don't block selection).
 *
 * Returns a closure-based checker: the dimension filter map is built once,
 * then each row is checked against it on demand (efficient for virtualized tables).
 */
export function computePivotRowSelection(
  config: PivotDataStoreConfig,
  _tableData: PivotDataRow[],
  dimensionFilters: Map<string, Set<string>>,
): PivotRowSelectionState {
  const hasActiveSelection = dimensionFilters.size > 0;

  const maxFilteredDimensionIndex = hasActiveSelection
    ? config.rowDimensionNames.reduce(
        (max, name, idx) => (dimensionFilters.has(name) ? idx : max),
        -1,
      )
    : -1;

  if (!hasActiveSelection) {
    return {
      hasActiveSelection: false,
      isRowSelected: () => false,
      maxFilteredDimensionIndex: -1,
    };
  }

  return {
    hasActiveSelection: true,
    maxFilteredDimensionIndex,
    isRowSelected: (
      rowData: PivotDataRow,
      depth?: number,
      parentRows?: PivotDataRow[],
    ) => {
      // In nested mode, all rows store their value under rowDimensions[0]
      // but the actual dimension is rowDimensionNames[depth]. When depth
      // is provided, use it for correct dimension mapping.
      if (depth !== undefined && !config.isFlat) {
        const firstDim = config.rowDimensionNames[0];
        const actualDim = config.rowDimensionNames[depth];
        const value = rowData[firstDim];
        if (value === undefined) return false;
        const strValue = String(value ?? "");
        if (!strValue) return false;

        const selectedSet = dimensionFilters.get(actualDim);
        if (selectedSet && !selectedSet.has(strValue)) return false;

        // Also verify all ancestor dimensions match their filters.
        // parentRows[i] is at depth i and stores its value under firstDim.
        if (parentRows) {
          for (let i = 0; i < parentRows.length; i++) {
            const ancestorDim = config.rowDimensionNames[i];
            const ancestorSet = dimensionFilters.get(ancestorDim);
            if (!ancestorSet) continue;
            const ancestorValue = String(parentRows[i][firstDim] ?? "");
            if (!ancestorValue || !ancestorSet.has(ancestorValue)) return false;
          }
        }

        return true;
      }

      const rowDimValues = getDimensionValuesFromRowData(config, rowData);
      if (rowDimValues.length === 0) return false;

      // A row is selected if every dimension that IS filtered
      // contains the row's value for that dimension
      return rowDimValues.every(({ dimensionName, value }) => {
        const selectedSet = dimensionFilters.get(dimensionName);
        if (!selectedSet) return true; // dimension not filtered; doesn't block
        return selectedSet.has(value);
      });
    },
  };
}

/**
 * Like getFiltersForCell but only extracts row dimension filters.
 * Used when clicking a row header (as opposed to a data cell).
 */
export function getFiltersForRowHeader(
  config: PivotDataStoreConfig,
  rowId: string,
  tableData: PivotDataRow[],
): PivotFilter {
  const values = getRawRowValues(config, rowId, tableData);
  const dimEntries = values.map((value, index) => ({
    name: config.rowDimensionNames[index],
    value,
  }));
  return buildPivotFilter(config, dimEntries);
}

/**
 * Builds filters for a column dimension header click.
 * Accepts the full dimension path (all ancestor + self dimension values)
 * so that clicking a nested column header filters on all parent dimensions too.
 */
export function getFiltersForColumnHeader(
  config: PivotDataStoreConfig,
  dimensionPath: Record<string, string>,
): PivotFilter {
  const dimEntries = Object.entries(dimensionPath).map(([name, value]) => ({
    name,
    value,
  }));
  return buildPivotFilter(config, dimEntries);
}

// --- Merged from pivot-filter-extraction.ts ---

export interface ExtractedFilter {
  dimensionName: string;
  values: (string | null)[];
}

/**
 * Extracts dimension filters from a V1Expression structure.
 * The expression is expected to be an AND operation containing IN operations.
 */
export function extractDimensionFiltersFromExpression(
  filters: V1Expression | undefined,
): ExtractedFilter[] {
  if (!filters?.cond?.exprs) return [];

  const result: ExtractedFilter[] = [];

  for (const expr of filters.cond.exprs) {
    if (expr.cond?.op === V1Operation.OPERATION_IN) {
      const ident = expr.cond.exprs?.[0]?.ident;
      if (!ident) continue;

      const values = expr?.cond?.exprs
        ?.slice(1)
        .map((e) => e.val)
        .filter((val): val is string | null => val !== undefined);

      if (values?.length) {
        result.push({ dimensionName: ident, values });
      }
    }
  }

  return result;
}

/**
 * Returns the set of dimension names present in a top-level AND expression.
 * Checks that each sub-expression is an IN operation before extracting the
 * ident, so LIKE/NIN/other filter types are not incorrectly included.
 *
 * Shared by pruneUnsub, excludeOwnDimensionFilters, and preExistingDims
 * to ensure consistent dimension detection across the codebase.
 */
export function getActiveDimensionNames(
  expr: V1Expression | undefined,
): Set<string> {
  if (!expr?.cond?.exprs) return new Set();
  const names = new Set<string>();
  for (const sub of expr.cond.exprs) {
    if (sub.cond?.op !== V1Operation.OPERATION_IN) continue;
    const ident = sub.cond.exprs?.[0]?.ident;
    if (ident) names.add(ident);
  }
  return names;
}
