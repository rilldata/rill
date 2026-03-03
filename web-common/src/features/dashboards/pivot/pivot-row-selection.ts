import { getValuesForExpandedKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-expansion";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  createAndExpression,
  createInExpression,
  getValuesInExpression,
} from "../stores/filter-utils";
import {
  getTimeForQuery,
  getTimeGrainFromDimension,
  getValuesForFlatTable,
  isTimeDimension,
} from "./pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
  PivotFilter,
  TimeFilters,
} from "./types";

/**
 * Tracks which rows were selected via row-header clicks and which
 * individual cells were selected via data-cell clicks. Supports
 * mixed multi-select: row-header clicks and cell clicks can coexist.
 */
export interface PivotClickSelectionState {
  /** rowIds selected via row-header clicks */
  rowHeaderSelections: Set<string>;
  /** "rowId:columnId" keys selected via data-cell clicks */
  cellSelections: Set<string>;
  /** Whether any selection exists at all */
  hasAnySelection: boolean;
  /** Check if a specific row was selected via row-header click */
  isRowHeaderSelected: (rowId: string) => boolean;
  /** Check if a specific cell was selected via data-cell click */
  isCellSelected: (rowId: string, columnId: string) => boolean;
}

export function createEmptyClickSelectionState(): PivotClickSelectionState {
  return {
    rowHeaderSelections: new Set(),
    cellSelections: new Set(),
    hasAnySelection: false,
    isRowHeaderSelected: () => false,
    isCellSelected: () => false,
  };
}

export function cellKey(rowId: string, columnId: string): string {
  return `${rowId}:${columnId}`;
}

export interface PivotRowSelectionState {
  /** True if at least one row dimension has active filter selections */
  hasActiveSelection: boolean;
  /** Check if a specific row (by rowId) is selected based on current filters */
  isRowSelected: (rowId: string) => boolean;
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
  const { rowDimensionNames, measureNames, isFlat } = config;

  let values: string[];
  if (isFlat) {
    values = getValuesForFlatTable(
      tableData,
      rowDimensionNames,
      rowId,
      measureNames.length > 0,
    );
  } else {
    values = getValuesForExpandedKey(
      tableData,
      rowDimensionNames,
      rowId,
      measureNames.length > 0,
    );
  }

  return values
    .map((value, index) => ({
      dimensionName: rowDimensionNames[index],
      value,
    }))
    .filter(
      ({ dimensionName }) =>
        !isTimeDimension(dimensionName, config.time.timeDimension),
    );
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
  tableData: PivotDataRow[],
  dimensionFilters: Map<string, Set<string>>,
): PivotRowSelectionState {
  const hasActiveSelection = dimensionFilters.size > 0;

  if (!hasActiveSelection) {
    return {
      hasActiveSelection: false,
      isRowSelected: () => false,
    };
  }

  return {
    hasActiveSelection: true,
    isRowSelected: (rowId: string) => {
      const rowDimValues = getDimensionValuesForRow(config, rowId, tableData);
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
  const { rowDimensionNames, measureNames, isFlat } = config;

  let values: string[];

  if (isFlat) {
    values = getValuesForFlatTable(
      tableData,
      rowDimensionNames,
      rowId,
      measureNames.length > 0,
    );
  } else {
    values = getValuesForExpandedKey(
      tableData,
      rowDimensionNames,
      rowId,
      measureNames.length > 0,
    );
  }

  const rowNestTimeFilters: TimeFilters[] = [];
  const rowNestFilters = values
    .map((value, index) =>
      createInExpression(rowDimensionNames[index], [value]),
    )
    .filter((f) => {
      if (
        isTimeDimension(f.cond?.exprs?.[0].ident, config.time.timeDimension)
      ) {
        rowNestTimeFilters.push({
          timeStart: f.cond?.exprs?.[1].val as string,
          interval: getTimeGrainFromDimension(
            f.cond?.exprs?.[0].ident as string,
          ),
        });
        return false;
      } else return true;
    });

  let rowFilters: V1Expression | undefined = undefined;
  if (rowNestFilters.length) {
    rowFilters = createAndExpression(rowNestFilters);
  }

  const timeRange: TimeRangeString = getTimeForQuery(
    config.time,
    rowNestTimeFilters,
  );

  const filters = mergeFilters(rowFilters, config.whereFilter);

  return { filters, timeRange };
}
