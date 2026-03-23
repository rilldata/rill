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

export interface PivotRowSelectionState {
  /** True if at least one row dimension has active filter selections */
  hasActiveSelection: boolean;
  /** Check if a specific row (by rowId) is selected based on current filters */
  isRowSelected: (rowId: string) => boolean;
  /** Highest row dimension index with an active selection filter, or -1 */
  maxFilteredDimensionIndex: number;
}

/**
 * Returns the raw dimension values for a row, including time dimensions.
 * Shared by getDimensionValuesForRow (which filters out time) and
 * getFiltersForRowHeader (which needs time for TimeFilters).
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
  const { rowDimensionNames } = config;
  const values = getRawRowValues(config, rowId, tableData);

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

/**
 * Builds filters for a column dimension header click.
 * Accepts the full dimension path (all ancestor + self dimension values)
 * so that clicking a nested column header filters on all parent dimensions too.
 */
export function getFiltersForColumnHeader(
  config: PivotDataStoreConfig,
  dimensionPath: Record<string, string>,
): PivotFilter {
  const defaultTimeRange = {
    start: config.time.timeStart,
    end: config.time.timeEnd,
  };

  const timeFilters: TimeFilters[] = [];
  const dimExprs = Object.entries(dimensionPath)
    .map(([name, value]) => {
      const expr = createInExpression(name, [value]);
      if (isTimeDimension(name, config.time.timeDimension)) {
        timeFilters.push({
          timeStart: value,
          interval: getTimeGrainFromDimension(name),
        });
        return null;
      }
      return expr;
    })
    .filter(Boolean) as V1Expression[];

  const timeRange =
    timeFilters.length > 0
      ? getTimeForQuery(config.time, timeFilters)
      : defaultTimeRange;

  const dimFilter =
    dimExprs.length > 0 ? createAndExpression(dimExprs) : undefined;
  const filters = mergeFilters(dimFilter, config.whereFilter);
  return { filters, timeRange };
}

// --- Merged from pivot-filter-extraction.ts ---

export interface ExtractedFilter {
  dimensionName: string;
  values: string[];
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
        .filter((val): val is string => val !== undefined && val !== null);

      if (values?.length) {
        result.push({ dimensionName: ident, values });
      }
    }
  }

  return result;
}
