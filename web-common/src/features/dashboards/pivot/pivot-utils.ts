import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import {
  Period,
  TimeOffsetType,
  TimeRangeString,
} from "@rilldata/web-common/lib/time/types";
import type {
  MetricsViewFilterCond,
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewAggregationSort,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { getColumnFiltersForPage } from "./pivot-infinite-scroll";
import type {
  PivotAxesData,
  PivotDataStoreConfig,
  PivotState,
  PivotTimeConfig,
  TimeFilters,
} from "./types";
import { PivotChipType } from "./types";
import type { PivotChipData } from "./types";

export function getMeasuresInPivotColumns(
  pivot: PivotState,
  measures: MetricsViewSpecMeasureV2[],
): string[] {
  const { columns } = pivot;

  return columns.filter(
    (rowName) => measures.findIndex((m) => m?.name === rowName) > -1,
  );
}

export function getDimensionsInPivotRow(
  pivot: PivotState,
  measures: MetricsViewSpecMeasureV2[],
): string[] {
  const { rows } = pivot;
  return rows.filter(
    (rowName) => measures.findIndex((m) => m?.name === rowName) === -1,
  );
}

export function getDimensionsInPivotColumns(
  pivot: PivotState,
  measures: MetricsViewSpecMeasureV2[],
): string[] {
  const { columns } = pivot;
  return columns.filter(
    (colName) => measures.findIndex((m) => m?.name === colName) === -1,
  );
}

export function getFormattedColumn(
  pivot: PivotState,
  allMeasures: MetricsViewSpecMeasureV2[],
  alldimensions: MetricsViewSpecDimensionV2[],
) {
  const measures: PivotChipData[] = [];
  const timeAndDimensions: PivotChipData[] = [];

  const { columns } = pivot;

  columns.forEach((colName) => {
    let label = "";
    let id = "";
    let chipType = PivotChipType.Measure;

    const measure = allMeasures.find((m) => m?.name === colName);

    if (measure && measure.label && measure.name) {
      chipType = PivotChipType.Measure;
      label = measure.label;
      id = measure.name;

      measures.push({ id, title: label, type: chipType });

      return;
    }

    const dimension = alldimensions.find((d) => d?.name === colName);

    if (dimension && dimension.label && dimension.name) {
      chipType = PivotChipType.Dimension;
      label = dimension.label;
      id = dimension.name;
    } else {
      chipType = PivotChipType.Time;
      label = colName;
      id = colName;
    }

    timeAndDimensions.push({ id, title: label, type: chipType });
  });

  return timeAndDimensions.concat(measures);
}

export function getFormattedRow(
  pivot: PivotState,
  allDimensions: MetricsViewSpecDimensionV2[],
) {
  const data: PivotChipData[] = [];

  const { rows } = pivot;

  rows.forEach((rowName) => {
    let label = "";
    let id = "";
    let chipType = PivotChipType.Dimension;

    const dimension = allDimensions.find((m) => m?.name === rowName);

    if (dimension && dimension.label && dimension.name) {
      chipType = PivotChipType.Dimension;
      label = dimension.label;
      id = dimension.name;
    } else {
      chipType = PivotChipType.Time;
      label = rowName;
      id = rowName;
    }

    data.push({ id, title: label, type: chipType });
  });

  return data;
}

export function getFormattedHeaderValues(
  pivot: PivotState,
  allMeasures: MetricsViewSpecMeasureV2[],
  alldimensions: MetricsViewSpecDimensionV2[],
) {
  const rows = getFormattedRow(pivot, alldimensions);
  const columns = getFormattedColumn(pivot, allMeasures, alldimensions);

  return {
    rows,
    columns,
  };
}

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
  const { colDimensionNames, rowDimensionNames, measureNames, filters, pivot } =
    config;

  const { sorting } = pivot;
  const sortingKey = JSON.stringify(sorting);
  const filterKey = JSON.stringify(filters);
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
) {
  // TODO: handle for already existing global filters

  const { colDimensionNames, rowDimensionNames, time } = config;

  let rowFilters: MetricsViewFilterCond[] = [];
  if (
    isInitialTable &&
    anchorDimension &&
    anchorDimension !== time.timeDimension
  ) {
    rowFilters = [
      {
        name: rowDimensionNames[0],
        in: rowDimensionValues.slice(0, yLimit),
      },
    ];
  }

  const colFiltersForPage = getColumnFiltersForPage(
    colDimensionNames.filter(
      (dimension) => dimension !== config.time.timeDimension,
    ),
    colDimensionAxes,
    config.pivot.columnPage,
    config.measureNames.length,
  );

  const filters = {
    include: [...colFiltersForPage, ...rowFilters],
    exclude: [],
  };

  return filters;
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
) {
  let sortPivotBy: V1MetricsViewAggregationSort[] = [];

  const defaultTimeRange = {
    start: config.time.timeStart,
    end: config.time.timeEnd,
  };

  // Return un-changed filter if no sorting is applied
  if (config.pivot?.sorting?.length === 0) {
    return {
      filters: config.filters,
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
      filters: config.filters,
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
      filters: config.filters,
      sortPivotBy,
      timeRange: defaultTimeRange,
    };
  }

  // Strip the measure string from the accessor
  const [accessorWithoutMeasure, measureIndex] = accessor.split("m");
  const accessorParts = accessorWithoutMeasure.split("_");

  let colDimensionFilters: MetricsViewFilterCond[];
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

        return {
          name: columnDimensionName,
          in: [value],
        };
      })
      .filter((colFilter) => {
        if (colFilter.name === config.time.timeDimension) {
          timeFilters.push({
            timeStart: colFilter.in[0],
            interval: config.time.interval,
          });
          return false;
        } else return true;
      });
  }

  const filterForSort: V1MetricsViewFilter = {
    include: [...colDimensionFilters],
    exclude: [],
  };

  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);

  sortPivotBy = [
    {
      desc: config.pivot.sorting[0].desc,
      name: measureNames[parseInt(measureIndex)],
    },
  ];

  return {
    filters: filterForSort,
    sortPivotBy,
    timeRange,
  };
}
