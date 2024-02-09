import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import {
  AvailableTimeGrain,
  Period,
  TimeOffsetType,
  TimeRangeString,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import { getColumnFiltersForPage } from "./pivot-infinite-scroll";
import type {
  PivotAxesData,
  PivotDataStoreConfig,
  PivotTimeConfig,
  TimeFilters,
} from "./types";

/**
 * Returns a sorted data array by appending the missing values in
 * sorted row axes data
 */
export function reconcileMissingDimensionValues(
  anchorDimension: string,
  sortedRowAxes: PivotAxesData | null,
  unsortedRowAxes: PivotAxesData | null,
) {
  // Return empty data if either sortedRowAxes or unsortedRowAxes is null
  if (!sortedRowAxes || !unsortedRowAxes) {
    return { rows: [], totals: [] };
  }

  // Extract data and totals from sortedRowAxes
  const sortedRowAxisValues = sortedRowAxes.data?.[anchorDimension] || [];
  const sortedTotals = sortedRowAxes.totals?.[anchorDimension] || [];

  // Return early if there are too many values
  if (sortedRowAxisValues.length >= 100) {
    return {
      rows: sortedRowAxisValues.slice(0, 100),
      totals: sortedTotals.slice(0, 100),
    };
  }

  // Extract data and totals from unsortedRowAxes
  const unsortedRowAxisValues = unsortedRowAxes.data?.[anchorDimension] || [];
  const unsortedTotals = unsortedRowAxes.totals?.[anchorDimension] || [];

  // Find missing values that are in unsortedRowAxes but not in sortedRowAxes
  const missingValues = unsortedRowAxisValues.filter(
    (value) => !sortedRowAxisValues.includes(value),
  );

  // Combine and limit the rows to 100
  const combinedRows = [...sortedRowAxisValues, ...missingValues].slice(0, 100);

  // Reorder the totals to match the order of combinedRows
  const reorderedTotals = combinedRows.map((rowValue) => {
    const sortedTotal = sortedTotals.find(
      (total) => total[anchorDimension] === rowValue,
    );
    if (sortedTotal) {
      return sortedTotal;
    }
    // Use the total from unsortedRowAxes if not found in sortedTotals
    const unsortedTotal = unsortedTotals.find(
      (total) => total[anchorDimension] === rowValue,
    );
    return unsortedTotal || { [anchorDimension]: rowValue };
  });

  return {
    rows: combinedRows,
    totals: reorderedTotals,
  };
}

/**
 * Construct a key for a pivot config to store expanded table data
 * in the cache
 */
export function getPivotConfigKey(config: PivotDataStoreConfig) {
  const {
    colDimensionNames,
    rowDimensionNames,
    measureNames,
    whereFilter,
    pivot,
  } = config;

  const { sorting } = pivot;
  const sortingKey = JSON.stringify(sorting);
  const filterKey = JSON.stringify(whereFilter);
  const dimsAndMeasures = rowDimensionNames
    .concat(measureNames, colDimensionNames)
    .join("_");

  return `${dimsAndMeasures}_${sortingKey}_${filterKey}`;
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
    // FIXME: Fix type warnings. Are these false positives?
    // Using `as` to avoid type warnings
    const duration: Period = TIME_GRAIN[filter.interval]?.duration as Period;

    const startTimeDt = new Date(filter.timeStart);
    const endTimeDt = getOffset(
      startTimeDt,
      duration,
      TimeOffsetType.ADD,
      timeZone,
    ) as Date;
    if (startTimeDt > new Date(timeStart as string)) {
      timeStart = filter.timeStart;
    }
    if (endTimeDt < new Date(timeEnd as string)) {
      timeEnd = endTimeDt.toISOString();
    }
  });

  return { start: timeStart, end: timeEnd };
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
export function getTotalColumnCount(
  columnDimensionAxes: Record<string, string[]> | undefined,
) {
  if (!columnDimensionAxes) return 0;

  return Object.values(columnDimensionAxes).reduce(
    (acc, columnDimension) => acc * columnDimension.length,
    1,
  );
}

/***
 * Get filter to be applied on aggregrate query for table cells
 */
export function getFilterForPivotTable(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  rowDimensionValues: string[] = [],
  isInitialTable = false,
  anchorDimension: string | undefined = undefined,
  yLimit = 100,
): V1Expression {
  // TODO: handle for already existing global filters

  const { colDimensionNames, rowDimensionNames, time } = config;

  let rowFilters: V1Expression | undefined;
  if (
    isInitialTable &&
    anchorDimension &&
    isTimeDimension(anchorDimension, time.timeDimension)
  ) {
    rowFilters = createInExpression(
      rowDimensionNames[0],
      rowDimensionValues.slice(0, yLimit),
    );
  }

  const colFiltersForPage = getColumnFiltersForPage(
    colDimensionNames.filter(
      (dimension) => !isTimeDimension(dimension, time.timeDimension),
    ),
    colDimensionAxes,
    config.pivot.columnPage,
    config.measureNames.length,
  );

  return createAndExpression([
    ...colFiltersForPage,
    ...(rowFilters ? [rowFilters] : []),
  ]);
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
  timeDimension: string,
  colValuesIndexMaps: Map<string, number>[],
  numMeasures: number,
  cell: { [key: string]: string | number },
) {
  const nestedColumnValueAccessor = colDimensionNames
    .map((colName, i) => {
      let accessor = `c${i}`;

      let colValue = cell[colName] as string;
      if (!colValue && isTimeDimension(colName, timeDimension)) {
        colValue = cell[timeDimension] as string;
      }

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
function extractNumbers(str: string) {
  const indexOfC = str.indexOf("c");
  const indexOfV = str.indexOf("v");

  const numberAfterC = parseInt(str.substring(indexOfC + 1, indexOfV));
  const numberAfterV = parseInt(str.substring(indexOfV + 1));

  return { c: numberAfterC, v: numberAfterV };
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
  where: V1Expression;
  sortPivotBy: V1MetricsViewAggregationSort[];
  timeRange: TimeRangeString;
} {
  let sortPivotBy: V1MetricsViewAggregationSort[] = [];

  const defaultTimeRange = {
    start: config.time.timeStart,
    end: config.time.timeEnd,
  };

  // Return un-changed filter if no sorting is applied
  if (config.pivot?.sorting?.length === 0) {
    return {
      where: config.whereFilter,
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  const { rowDimensionNames, colDimensionNames, measureNames } = config;
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
      where: config.whereFilter,
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
      where: config.whereFilter,
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  // Strip the measure string from the accessor
  const [accessorWithoutMeasure, measureIndex] = accessor.split("m");
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

  const filterForSort: V1Expression = createAndExpression(colDimensionFilters);
  const mergedFilter = mergeFilters(config.whereFilter, filterForSort);

  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);

  sortPivotBy = [
    {
      desc: config.pivot.sorting[0].desc,
      name: measureNames[parseInt(measureIndex)],
    },
  ];

  return {
    where: mergedFilter,
    sortPivotBy,
    timeRange,
  };
}
