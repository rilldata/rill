import type {
  MetricsViewFilterCond,
  MetricsViewSpecMeasureV2,
  V1MetricsViewAggregationSort,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
  PivotState,
  PivotTimeConfig,
  TimeFilters,
} from "./types";
import type { ColumnDef } from "@tanstack/svelte-table";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  Period,
  TimeOffsetType,
  TimeRangeString,
} from "@rilldata/web-common/lib/time/types";

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

/**
 * Returns a sorted data array by appending the missing values in
 * sorted row axes data
 */
export function reconcileMissingDimensionValues(
  anchorDimension: string,
  sortedRowAxesData: Record<string, string[]> | undefined,
  unsortedRowAxesData: Record<string, string[]> | undefined,
) {
  const sortedRowAxisValues = new Set(
    sortedRowAxesData?.[anchorDimension] || [],
  );
  const unsortedRowAxisValues = unsortedRowAxesData?.[anchorDimension] || [];

  const missingValues = unsortedRowAxisValues.filter(
    (value) => !sortedRowAxisValues.has(value),
  );

  return [...sortedRowAxisValues, ...missingValues];
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

/***
 * Get filter for table cells
 */
export function getFilterForPivotTable(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  rowDimensionValues: string[] = [],
  isInitialTable = false,
  yLimit = 100,
  xLimit = 100,
) {
  // TODO: handle for already existing global filters

  const { colDimensionNames, rowDimensionNames, time } = config;

  let rowFilters: MetricsViewFilterCond[] = [];
  const anchorDimension = rowDimensionNames?.[0];
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
  const colFilters = colDimensionNames
    .filter((dimension) => dimension !== config.time.timeDimension)
    .map((colDimensionName) => {
      return {
        name: colDimensionName,
        in: colDimensionAxes?.[colDimensionName].slice(0, xLimit),
      };
    });

  const filters = {
    include: [...colFilters, ...rowFilters],
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

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  dimensionNames: string[],
  headers: Record<string, string[]>,
  leafData: ColumnDef<PivotDataRow>[],
): ColumnDef<PivotDataRow>[] {
  const colValuesIndexMaps = dimensionNames?.map((colDimension) =>
    createIndexMap(headers[colDimension]),
  );

  const levels = dimensionNames.length;
  // Recursive function to create nested columns
  function createNestedColumns(
    level: number,
    colValuePair: { [key: string]: string },
  ): ColumnDef<PivotDataRow>[] {
    if (level === levels) {
      const accessors = getAccessorForCell(
        dimensionNames,
        colValuesIndexMaps,
        leafData.length,
        colValuePair,
      );

      // Base case: return leaf columns
      return leafData.map((leaf, i) => ({
        ...leaf,
        // Change accessor key to match the nested column structure
        accessorKey: accessors[i],
      }));
    }

    // Recursive case: create nested headers
    const headerValues = headers[dimensionNames?.[level]];
    return headerValues?.map((value) => ({
      header: value,
      columns: createNestedColumns(level + 1, {
        ...colValuePair,
        [dimensionNames[level]]: value,
      }),
    }));
  }

  // Start the recursion
  return createNestedColumns(0, {});
}

export function getColumnDefForPivot(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> | undefined,
) {
  const IsNested = true;

  const { measureNames, rowDimensionNames, colDimensionNames } = config;

  const measures = measureNames.map((m) => ({
    label: config.allMeasures.find((measure) => measure.name === m)?.label || m,
    name: m,
  }));

  const rowDimensions = rowDimensionNames.map((d) => ({
    label:
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d,
    name: d,
  }));
  const colDimensions = colDimensionNames.map((d) => ({
    label:
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d,
    name: d,
  }));

  let rowDimensionsForColumnDef = rowDimensions;
  let nestedLabel: string;
  if (IsNested) {
    rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
    nestedLabel = rowDimensions.map((d) => d.label || d.name).join(" > ");
  }
  const rowDefinitions: ColumnDef<PivotDataRow>[] =
    rowDimensionsForColumnDef.map((d) => {
      return {
        accessorKey: d.name,
        header: nestedLabel,
        cell: ({ row, getValue }) =>
          cellComponent(PivotExpandableCell, {
            value: getValue(),
            row,
          }),
      };
    });

  const leafColumns: ColumnDef<PivotDataRow>[] = measures.map((m) => {
    return {
      accessorKey: m.name,
      header: m.label || m.name,
      cell: (info) => info.getValue(),
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    colDimensions.map((d) => d.name) || [],
    columnDimensionAxes || {},
    leafColumns,
  );

  return [...rowDefinitions, ...groupedColDef];
}
