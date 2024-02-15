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
  PivotDataRow,
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
    measureFilter,
    pivot,
  } = config;

  const { sorting } = pivot;
  const sortingKey = JSON.stringify(sorting);
  const filterKey = JSON.stringify(whereFilter);
  const measureFilterKey = JSON.stringify(measureFilter);
  const dimsAndMeasures = rowDimensionNames
    .concat(measureNames, colDimensionNames)
    .join("_");

  return `${dimsAndMeasures}_${sortingKey}_${filterKey}_${measureFilterKey}`;
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
  isInitialTable = false,
  anchorDimension: string | undefined = undefined,
  yLimit = 100,
) {
  // TODO: handle for already existing global filters

  const { rowDimensionNames, time } = config;

  let rowFilters: V1Expression | undefined;
  if (
    isInitialTable &&
    anchorDimension &&
    !isTimeDimension(anchorDimension, time.timeDimension)
  ) {
    rowFilters = createInExpression(
      rowDimensionNames[0],
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
  timeDimension: string,
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

  // Return un-changed filter if no sorting is applied
  if (config.pivot?.sorting?.length === 0) {
    return {
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

  let filterForSort: V1Expression | undefined;
  if (colDimensionFilters.length) {
    filterForSort = createAndExpression(colDimensionFilters);
  }
  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);

  sortPivotBy = [
    {
      desc: config.pivot.sorting[0].desc,
      name: measureNames[parseInt(measureIndex)],
    },
  ];

  return {
    where: filterForSort,
    sortPivotBy,
    timeRange,
  };
}
