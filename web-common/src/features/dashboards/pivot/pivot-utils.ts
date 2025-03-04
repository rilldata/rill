import { getValuesForExpandedKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-expansion";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import {
  type AvailableTimeGrain,
  Period,
  TimeOffsetType,
  type TimeRangeString,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { getColumnFiltersForPage } from "./pivot-infinite-scroll";
import { mergeFilters } from "./pivot-merge-filters";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
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

  const { sorting, tableMode: tableModeKey } = pivot;
  const timeKey = JSON.stringify(time);
  const sortingKey = JSON.stringify(sorting);
  const filterKey = JSON.stringify(whereFilter);
  const comparisonTimeKey = JSON.stringify(comparisonTime);
  const dimsAndMeasures = rowDimensionNames
    .concat(measureNames, colDimensionNames)
    .join("_");

  return `${dimsAndMeasures}_${timeKey}_${sortingKey}_${tableModeKey}_${filterKey}_${enableComparison}_${comparisonTimeKey}`;
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

function getColumnFiltersFromMinimizedAccessor(
  config: PivotDataStoreConfig,
  accessor: string,
  columnDimensionAxes: Record<string, string[]> = {},
) {
  const { colDimensionNames } = config;

  // Strip the measure string from the accessor
  const [accessorWithoutMeasure] = accessor.split("m");
  const accessorParts = accessorWithoutMeasure.split("_");

  let colDimensionFilters: V1Expression[];
  const timeFilters: TimeFilters[] = [];
  if (accessorParts[0] === "") {
    // There are no column dimensions in the accessor
    colDimensionFilters = [];
  } else {
    colDimensionFilters = accessorParts
      .map((part) => {
        const { c, v } = extractNumbers(part);
        const columnDimensionName = colDimensionNames[c];
        const value = columnDimensionAxes[columnDimensionName][v];

        return createInExpression(columnDimensionName, [value]);
      })
      .filter((colFilter) => {
        if (
          isTimeDimension(
            colFilter.cond?.exprs?.[0].ident,
            config.time.timeDimension,
          )
        ) {
          timeFilters.push({
            timeStart: colFilter.cond?.exprs?.[1].val as string,
            interval: getTimeGrainFromDimension(
              colFilter.cond?.exprs?.[0].ident as string,
            ),
          });
          return false;
        } else return true;
      });
  }

  let filterForSort: V1Expression | undefined;

  if (colDimensionFilters.length) {
    filterForSort = createAndExpression(colDimensionFilters);
  }
  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);
  return {
    filters: filterForSort,
    timeRange,
  };
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

  sortPivotBy = [
    {
      desc: config.pivot.sorting[0].desc,
      name: measureNames[parseInt(measureIndex)],
    },
  ];

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
  if (pivotState.columns.measure.length > 10) {
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

function getValuesForFlatTable(
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

export function getFiltersForCell(
  config: PivotDataStoreConfig,
  rowId: string,
  colId: string,
  colDimensionAxes: Record<string, string[]> = {},
  tableData: PivotDataRow[],
): PivotFilter {
  const { rowDimensionNames, measureNames, isFlat } = config;
  const defaultTimeRange = {
    start: config.time.timeStart,
    end: config.time.timeEnd,
  };

  let values: string[];

  if (isFlat) {
    // TODO: Update this when columns can be mixed with measures
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

  const timeRangeRow: TimeRangeString = getTimeForQuery(
    config.time,
    rowNestTimeFilters,
  );

  // Get filters for column dimensions
  let columnFilters: V1Expression | undefined;
  let timeRangeCol: TimeRangeString;
  const firstDimension = rowDimensionNames?.[0];
  if (firstDimension === colId || measureNames.includes(colId) || isFlat) {
    columnFilters = undefined;
    timeRangeCol = defaultTimeRange;
  } else {
    const { filters, timeRange } = getColumnFiltersFromMinimizedAccessor(
      config,
      colId,
      colDimensionAxes,
    );
    columnFilters = filters;
    timeRangeCol = timeRange;
  }

  const timeRange = mergeTimeStrings(timeRangeRow, timeRangeCol);
  const cellFilters = mergeFilters(rowFilters, columnFilters);
  const filters = mergeFilters(cellFilters, config.whereFilter);

  return { filters, timeRange };
}

export function getErrorFromResponse(
  queryResult: QueryObserverResult<V1MetricsViewAggregationResponse, HTTPError>,
): PivotQueryError {
  return {
    statusCode: queryResult?.error?.response?.status || null,
    message: queryResult?.error?.response?.data?.message,
  };
}

export function getErrorFromResponses(
  queryResults: (QueryObserverResult<
    V1MetricsViewAggregationResponse,
    HTTPError
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
