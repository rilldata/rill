import { getValuesForExpandedKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-expansion";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import {
  Period,
  TimeOffsetType,
  type AvailableTimeGrain,
  type TimeRangeString,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import { connectCodeToHTTPStatus } from "@rilldata/web-common/lib/errors";
import type { ConnectError } from "@connectrpc/connect";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { Row } from "tanstack-table-8-svelte-5";
import { SHOW_MORE_BUTTON } from "./pivot-constants";
import { getColumnFiltersForPage } from "./pivot-infinite-scroll";
import { mergeFilters } from "./pivot-merge-filters";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotChipData,
  type PivotDataRow,
  type PivotDataState,
  type PivotDataStoreConfig,
  type PivotFilter,
  type PivotQueryError,
  type PivotState,
  type PivotTimeConfig,
  type TimeFilters,
} from "./types";

/**
 * Construct a key for a pivot config to store expanded table data
 * in the cache
 */
export function getPivotConfigKey(config: PivotDataStoreConfig) {
  const {
    time,
    colDimensionNames,
    rowDimensionNames,
    measureNames,
    whereFilter,
    enableComparison,
    comparisonTime,
    pivot,
  } = config;

  const {
    sorting,
    tableMode: tableModeKey,
    rowLimit,
    outermostRowLimit,
  } = pivot;
  const timeKey = JSON.stringify(time);
  const sortingKey = JSON.stringify(sorting);
  const filterKey = JSON.stringify(whereFilter);
  const comparisonTimeKey = JSON.stringify(comparisonTime);
  const dimsAndMeasures = rowDimensionNames
    .concat(measureNames, colDimensionNames)
    .join("_");

  return `${dimsAndMeasures}_${timeKey}_${sortingKey}_${tableModeKey}_${filterKey}_${enableComparison}_${comparisonTimeKey}_${rowLimit ?? "all"}_${outermostRowLimit ?? "none"}`;
}

/**
 * Apply the time filters on global start and end time to get the
 * start and end time for the query
 */
export function getTimeForQuery(
  time: PivotTimeConfig,
  timeFilters: TimeFilters[],
): TimeRangeString {
  let { timeStart, timeEnd } = time;
  const { timeZone } = time;

  if (!timeStart || !timeEnd) {
    return { start: timeStart, end: timeEnd };
  }

  timeFilters.forEach((filter) => {
    const startTimeDt = new Date(filter.timeStart);
    let startTimeOfLastInterval: Date | undefined = undefined;

    if (filter.timeEnd) {
      startTimeOfLastInterval = new Date(filter.timeEnd);
    } else {
      startTimeOfLastInterval = startTimeDt;
    }

    const duration = TIME_GRAIN[filter.interval]?.duration as Period;
    const endTimeDt = getOffset(
      startTimeOfLastInterval,
      duration,
      TimeOffsetType.ADD,
      timeZone,
    );

    if (startTimeDt > new Date(timeStart as string)) {
      timeStart = filter.timeStart;
    }
    if (endTimeDt < new Date(timeEnd as string)) {
      timeEnd = endTimeDt.toISOString();
    }
  });

  return { start: timeStart, end: timeEnd };
}

/**
 * Returns the intersection of two time ranges
 */
export function mergeTimeStrings(
  time1: TimeRangeString,
  time2: TimeRangeString,
): TimeRangeString {
  if (!time1.start || !time1.end) {
    return time2;
  }
  if (!time2.start || !time2.end) {
    return time1;
  }

  const start1 = new Date(time1.start);
  const start2 = new Date(time2.start);
  const end1 = new Date(time1.end);
  const end2 = new Date(time2.end);

  const start = start1 > start2 ? start1 : start2;
  const end = end1 < end2 ? end1 : end2;

  return {
    start: start.toISOString(),
    end: end.toISOString(),
  };
}

export function isTimeDimension(
  dimension: string | undefined,
  timeDimension: string,
) {
  if (!dimension) return false;
  return dimension.startsWith(`${timeDimension}_rill_`);
}

export function getTimeGrainFromDimension(dimension: string) {
  const grainLabel = dimension.split("_rill_")[1];
  return grainLabel as AvailableTimeGrain;
}

/**
 * Alternative to flexRender for performant rendering of cells
 */
export const cellComponent = (
  component: unknown,
  props: Record<string, unknown>,
) => ({
  component,
  props,
});

/**
 * Create a value to index map for a given array
 */
export function createIndexMap<T>(arr: T[]): Map<T, number> {
  const indexMap = new Map<T, number>();
  arr.forEach((element, index) => {
    indexMap.set(element, index);
  });
  return indexMap;
}

/**
 * Returns total number of columns for the table
 * excluding row and group totals columns
 */
export function getTotalColumnCount(totalsRow: PivotDataRow) {
  return Object.keys(totalsRow).length;
}

/***
 * Get filter to be applied on aggregrate query for table cells
 */
export function getFilterForPivotTable(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  totalsRow: PivotDataRow,
  rowDimensionValues: string[] = [],
  anchorDimension: string | undefined = undefined,
  yLimit = 100,
) {
  const { isFlat, time } = config;

  let rowFilters: V1Expression | undefined;
  if (
    anchorDimension &&
    !isFlat &&
    !isTimeDimension(anchorDimension, time.timeDimension)
  ) {
    rowFilters = createInExpression(
      anchorDimension,
      rowDimensionValues.slice(0, yLimit),
    );
  }

  const { filters: colFiltersForPage, timeFilters } = getColumnFiltersForPage(
    config,
    colDimensionAxes,
    totalsRow,
  );

  const filters = createAndExpression([
    ...colFiltersForPage,
    ...(rowFilters ? [rowFilters] : []),
  ]);

  return { filters, timeFilters };
}

/**
 * Create a nested accessor for a cell in the table.
 * This is used to map the cell data to the table data.
 *
 * Column names are converted to c0, c1, c2, etc.
 * Column values are converted to v0, v1, v2, etc.
 * Measure names are converted to m0, m1, m2, etc.
 */
export function getAccessorForCell(
  colDimensionNames: string[],
  colValuesIndexMaps: Map<string, number>[],
  numMeasures: number,
  cell: { [key: string]: string | number },
) {
  const nestedColumnValueAccessor = colDimensionNames
    .map((colName, i) => {
      let accessor = `c${i}`;

      const colValue = cell[colName] as string;
      const colValueIndex = colValuesIndexMaps[i].get(colValue);
      accessor += `v${colValueIndex}`;

      return accessor;
    })
    .join("_");

  return Array(numMeasures)
    .fill(null)
    .map((_, i) => `${nestedColumnValueAccessor}m${i}`);
}

/**
 * Extract the numbers after c and v in a accessor part string
 */
export function extractNumbers(str: string) {
  const indexOfC = str.indexOf("c");
  const indexOfV = str.indexOf("v");

  const numberAfterC = parseInt(str.substring(indexOfC + 1, indexOfV));
  const numberAfterV = parseInt(str.substring(indexOfV + 1));

  return { c: numberAfterC, v: numberAfterV };
}

export function sortAcessors(accessors: string[]) {
  function parseParts(str: string): number[] {
    // Extract all occurrences of patterns like c<num>v<num>
    const matches = str.match(/c(\d+)v(\d+)/g);
    if (!matches) {
      return [];
    }
    // Map each found pattern to its numeric components
    const parts: number[] = matches.flatMap((match) => {
      const result = /c(\d+)v(\d+)/.exec(match);
      if (!result) return [];
      const [, cPart, vPart] = result;
      return [parseInt(cPart, 10), parseInt(vPart, 10)]; // Convert to numbers for proper comparison
    });

    // Extract m<num> part
    const mPartMatch = str.match(/m(\d+)$/);
    if (mPartMatch) {
      parts.push(parseInt(mPartMatch[1], 10)); // Add m<num> part as a number
    }
    return parts;
  }

  return accessors.sort((a: string, b: string): number => {
    const partsA = parseParts(a);
    const partsB = parseParts(b);

    // Compare each part until a difference is found
    for (let i = 0; i < Math.max(partsA.length, partsB.length); i++) {
      const partA = partsA[i] || 0; // Default to 0 if undefined
      const partB = partsB[i] || 0; // Default to 0 if undefined
      if (partA !== partB) {
        return partA - partB;
      }
    }

    // If all parts are equal, consider them equal
    return 0;
  });
}

/**
 * Extract column dimension name/value pairs from a minimized accessor string.
 * Used by getFiltersForCell/getFiltersFromRow to combine with row dim entries
 * before passing to buildPivotFilter.
 */
function getColumnDimEntriesFromAccessor(
  config: PivotDataStoreConfig,
  accessor: string,
  columnDimensionAxes: Record<string, string[]> = {},
): Array<{ name: string; value: string }> {
  const { colDimensionNames } = config;
  const [accessorWithoutMeasure] = accessor.split("m");
  const accessorParts = accessorWithoutMeasure.split("_");

  if (accessorParts[0] === "") return [];

  return accessorParts.map((part) => {
    const { c, v } = extractNumbers(part);
    const name = colDimensionNames[c];
    const value = columnDimensionAxes[name][v];
    return { name, value };
  });
}

/**
 * Legacy wrapper: returns column filters as V1Expression + timeRange.
 * Still used by sorting and other non-click-to-filter code paths.
 */
function getColumnFiltersFromMinimizedAccessor(
  config: PivotDataStoreConfig,
  accessor: string,
  columnDimensionAxes: Record<string, string[]> = {},
) {
  const entries = getColumnDimEntriesFromAccessor(
    config,
    accessor,
    columnDimensionAxes,
  );

  const timeFilters: TimeFilters[] = [];
  const dimExprs: V1Expression[] = [];

  for (const { name, value } of entries) {
    if (isTimeDimension(name, config.time.timeDimension)) {
      timeFilters.push({
        timeStart: value,
        interval: getTimeGrainFromDimension(name),
      });
    } else {
      dimExprs.push(createInExpression(name, [value]));
    }
  }

  const filterForSort =
    dimExprs.length > 0 ? createAndExpression(dimExprs) : undefined;
  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);
  return { filters: filterForSort, timeRange };
}

/**
 * Returns column dimension entries for a cell click, or an empty array
 * when the column is a row header dimension, a measure, or the table is flat.
 */
function getColumnDimEntries(
  config: PivotDataStoreConfig,
  colId: string,
  colDimensionAxes: Record<string, string[]>,
): Array<{ name: string; value: string }> {
  const { rowDimensionNames, measureNames, isFlat } = config;
  const firstDimension = rowDimensionNames?.[0];
  if (firstDimension === colId || measureNames.includes(colId) || isFlat) {
    return [];
  }
  return getColumnDimEntriesFromAccessor(config, colId, colDimensionAxes);
}

/**
 * For a given accessor created by getAccessorForCell, get the filter
 * that can be applied to the table to get sorted data based on the
 * accessor.
 */
export function getSortForAccessor(
  anchorDimension: string,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): {
  where?: V1Expression;
  sortPivotBy: V1MetricsViewAggregationSort[];
  timeRange: TimeRangeString;
} {
  let sortPivotBy: V1MetricsViewAggregationSort[] = [];

  const defaultTimeRange = {
    start: config.time.timeStart,
    end: config.time.timeEnd,
  };

  // Return un-changed filter if no sorting is applied or in flat mode
  if (config.pivot?.sorting?.length === 0 || config.isFlat) {
    return {
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  const { rowDimensionNames, measureNames } = config;
  const accessor = config.pivot.sorting[0].id;

  // For the first column, the accessor is the row dimension name
  const firstDimension = rowDimensionNames?.[0];
  if (firstDimension === accessor) {
    sortPivotBy = [
      {
        desc: config.pivot.sorting[0].desc,
        name: anchorDimension,
      },
    ];
    return {
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  // For the row totals, the accessor is the measure name
  if (measureNames.includes(accessor)) {
    sortPivotBy = [
      {
        desc: config.pivot.sorting[0].desc,
        name: accessor,
      },
    ];
    return {
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  const measureIndex = accessor.split("m")[1];
  const { filters, timeRange } = getColumnFiltersFromMinimizedAccessor(
    config,
    accessor,
    columnDimensionAxes,
  );

  const measureName = measureNames[parseInt(measureIndex)];

  if (measureName) {
    sortPivotBy = [
      {
        desc: config.pivot.sorting[0].desc,
        name: measureName,
      },
    ];
  }

  return {
    where: filters,
    sortPivotBy,
    timeRange,
  };
}

export function getFilterForMeasuresTotalsAxesQuery(
  config: PivotDataStoreConfig,
  anchorDimension: string,
  rowDimensionValues: string[],
): V1Expression | undefined {
  if (isTimeDimension(anchorDimension, config.time.timeDimension)) {
    return config.whereFilter;
  }
  const rowFilters = createAndExpression([
    createInExpression(anchorDimension, rowDimensionValues),
  ]);
  const mergedFilters = mergeFilters(rowFilters, config.whereFilter);

  return mergedFilters;
}

export function prepareMeasureForComparison(
  measures: V1MetricsViewAggregationMeasure[],
): V1MetricsViewAggregationMeasure[] {
  return measures.map((measure) => {
    if (measure.name?.endsWith(COMPARISON_PERCENT)) {
      return {
        ...measure,
        comparisonRatio: {
          measure: measure.name.replace(COMPARISON_PERCENT, ""),
        },
      };
    } else if (measure.name?.endsWith(COMPARISON_DELTA)) {
      return {
        ...measure,
        comparisonDelta: {
          measure: measure.name.replace(COMPARISON_DELTA, ""),
        },
      };
    }

    return measure;
  });
}

export function canEnablePivotComparison(
  pivotState: PivotState,
  comparisonStart: string | Date | undefined,
) {
  // Disable if more than 10 measures

  const measures = splitPivotChips(pivotState.columns).measure;
  if (measures.length > 10) {
    return false;
  }
  // Disable if time comparison is not present
  if (!comparisonStart) {
    return false;
  }

  return true;
}

export function getSortFilteredMeasureBody(
  measureBody: V1MetricsViewAggregationMeasure[],
  sortPivotBy: V1MetricsViewAggregationSort[],
  measureWhere: V1Expression | undefined,
) {
  let sortFilteredMeasureBody: V1MetricsViewAggregationMeasure[] = measureBody;
  let isMeasureSortAccessor = false;
  let sortAccessor: string | undefined = undefined;

  if (sortPivotBy.length && measureWhere) {
    sortAccessor = sortPivotBy[0]?.name;

    isMeasureSortAccessor = measureBody.some((m) => m.name === sortAccessor);
    if (isMeasureSortAccessor && sortAccessor) {
      sortFilteredMeasureBody = [{ name: sortAccessor, filter: measureWhere }];
    }
  }

  return { sortFilteredMeasureBody, isMeasureSortAccessor, sortAccessor };
}

export function getValuesForFlatTable(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  rowId: string,
  hasTotalsRow: boolean,
): string[] {
  let index = parseInt(rowId, 10);
  const dimensionValues: string[] = [];

  if (hasTotalsRow) index = index - 1;

  const row = tableData?.[index];
  if (!row) return dimensionValues;

  // For flat tables, collect all dimension values in order
  rowDimensions.forEach((dim) => {
    if (dim in row) {
      dimensionValues.push(row[dim] as string);
    }
  });

  return dimensionValues;
}

/**
 * Shared core for all pivot filter builders. Takes dimension name/value pairs,
 * separates time dimensions into TimeFilters, creates IN expressions for the rest,
 * computes the narrowed time range, and merges everything with optional extra filters.
 *
 * Every public getFiltersFor* function delegates here after extracting its
 * dimension entries from whichever data source it uses (positional rowId,
 * direct rowData, column header path, etc.).
 */
export function buildPivotFilter(
  config: PivotDataStoreConfig,
  dimEntries: Array<{ name: string; value: string | null }>,
  extraFilters?: V1Expression,
): PivotFilter {
  const timeFilters: TimeFilters[] = [];
  const dimExprs: V1Expression[] = [];

  for (const { name, value } of dimEntries) {
    const expr = createInExpression(name, [value]);
    if (isTimeDimension(name, config.time.timeDimension)) {
      timeFilters.push({
        timeStart: value as string,
        interval: getTimeGrainFromDimension(name),
      });
    } else {
      dimExprs.push(expr);
    }
  }

  const dimFilter =
    dimExprs.length > 0 ? createAndExpression(dimExprs) : undefined;
  const timeRange = getTimeForQuery(config.time, timeFilters);

  let filters: V1Expression | undefined;
  if (extraFilters) {
    const combined = mergeFilters(dimFilter, extraFilters);
    filters = mergeFilters(combined, config.whereFilter);
  } else {
    filters = mergeFilters(dimFilter, config.whereFilter);
  }

  return { filters, timeRange };
}

export function getFiltersForCell(
  config: PivotDataStoreConfig,
  rowId: string,
  colId: string,
  colDimensionAxes: Record<string, string[]> = {},
  tableData: PivotDataRow[],
  upToDimensionIndex?: number,
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
    if (upToDimensionIndex !== undefined && upToDimensionIndex >= 0) {
      values = values.slice(0, upToDimensionIndex + 1);
    }
  } else {
    values = getValuesForExpandedKey(
      tableData,
      rowDimensionNames,
      rowId,
      measureNames.length > 0,
    );
  }

  const rowEntries = values.map((value, index) => ({
    name: rowDimensionNames[index],
    value,
  }));

  const colEntries = getColumnDimEntries(config, colId, colDimensionAxes);
  return buildPivotFilter(config, [...rowEntries, ...colEntries]);
}

/**
 * Like getFiltersForCell but reads dimension values directly from a
 * PivotDataRow object instead of doing parseInt(rowId) array indexing.
 * This makes selections stable across sorting and data refreshes.
 */
export function getFiltersFromRow(
  config: PivotDataStoreConfig,
  rowData: PivotDataRow,
  colId: string,
  colDimensionAxes: Record<string, string[]> = {},
  upToDimensionIndex?: number,
): PivotFilter {
  let dimNames = config.rowDimensionNames;
  if (upToDimensionIndex !== undefined && upToDimensionIndex >= 0) {
    dimNames = config.rowDimensionNames.slice(0, upToDimensionIndex + 1);
  }

  const rowEntries: Array<{ name: string; value: string | null }> = [];
  for (const dim of dimNames) {
    const val = rowData[dim];
    if (val === undefined) continue;
    if (val === null) {
      rowEntries.push({ name: dim, value: null });
    } else if (typeof val === "string" || typeof val === "number") {
      rowEntries.push({ name: dim, value: String(val) });
    }
  }

  const colEntries = getColumnDimEntries(config, colId, colDimensionAxes);
  return buildPivotFilter(config, [...rowEntries, ...colEntries]);
}

export function getErrorFromResponse(
  queryResult: QueryObserverResult<
    V1MetricsViewAggregationResponse,
    ConnectError
  >,
): PivotQueryError {
  const err = queryResult?.error;
  const statusCode = err ? connectCodeToHTTPStatus(err.code) : null;
  const message = err?.rawMessage || err?.message;
  return { statusCode, message };
}

export function getErrorFromResponses(
  queryResults: (QueryObserverResult<
    V1MetricsViewAggregationResponse,
    ConnectError
  > | null)[],
): PivotQueryError[] {
  return queryResults
    .filter((result) => result?.isError)
    .map(getErrorFromResponse);
}

export function getErrorState(errors: PivotQueryError[]): PivotDataState {
  return {
    error: errors,
    isFetching: false,
    data: [],
    columnDef: [],
    assembled: false,
    totalColumns: 0,
  };
}

export function isElement(target: EventTarget | null): target is HTMLElement {
  return target instanceof HTMLElement;
}

export function splitPivotChips(data: PivotChipData[]): {
  dimension: PivotChipData[];
  measure: PivotChipData[];
} {
  return {
    dimension: data?.filter((c) => c.type !== PivotChipType.Measure) || [],
    measure: data?.filter((c) => c.type === PivotChipType.Measure) || [],
  };
}

/**
 * Check if a row is a "show more" button row by inspecting the first cell value
 */
export function isShowMoreRow(row: Row<PivotDataRow>): boolean {
  const firstCell = row?.getVisibleCells()?.[0];
  return firstCell?.getValue() === SHOW_MORE_BUTTON;
}
