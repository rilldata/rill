import type {
  MetricsViewFilterCond,
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import type { PivotState } from "./types";
import { getAccessorForCell } from "./pivot-table-transformations";

export function getMeasuresInPivotColumns(
  pivot: PivotState,
  measures: MetricsViewSpecMeasureV2[]
): MetricsViewSpecMeasureV2[] {
  const { columns } = pivot;

  return columns
    .filter((rowName) => measures.findIndex((m) => m?.name === rowName) > -1)
    .map((rowName) => measures.find((m) => m?.name === rowName));
}

export function getDimensionsInPivotRow(
  pivot: PivotState,
  dimensions: MetricsViewSpecDimensionV2[]
): MetricsViewSpecDimensionV2[] {
  const { rows } = pivot;
  return rows
    .filter(
      (rowName) => dimensions.findIndex((m) => m?.column === rowName) > -1
    )
    .map((rowName) => dimensions.find((m) => m?.column === rowName));
}

export function getDimensionsInPivotColumns(
  pivot: PivotState,
  dimensions: MetricsViewSpecDimensionV2[]
): MetricsViewSpecDimensionV2[] {
  const { columns } = pivot;
  return columns
    .filter(
      (colName) => dimensions.findIndex((m) => m?.column === colName) > -1
    )
    .map((colName) => dimensions.find((m) => m?.column === colName));
}

/**
 * At the start we don't have enough information about the values present in an expanded group
 * For now fill it with empty values if there are more than one row dimensions
 */
export function prepareExpandedPivotData(
  data,
  dimensions: string[],
  expanded,
  i = 1
) {
  if (dimensions.slice(i).length > 0) {
    data.forEach((row) => {
      row.subRows = [{ [dimensions[0]]: "LOADING_CELL" }];

      prepareExpandedPivotData(row.subRows, dimensions, expanded, i + 1);
    });
  }
}

/**
 * Alternative to flexRender for performant rendering of cells
 */
export const cellComponent = (
  component: unknown,
  props: Record<string, unknown>
) => ({
  component,
  props,
});

/**
 * Create a value to index map for a given array
 */
export function createIndexMap(arr) {
  const indexMap = new Map();
  arr.forEach((element, index) => {
    indexMap.set(element, index);
  });
  return indexMap;
}

/***
 * Get filter for table cells
 */
export function getFilterForPivotTable(
  config,
  colDimensionAxes,
  rowDimensionAxes,
  yLimit = 100,
  xLimit = 20
) {
  // TODO: handle for already existing global filters

  const { colDimensionNames, rowDimensionNames } = config;

  let rowFilters: MetricsViewFilterCond[] = [];
  if (rowDimensionNames.length) {
    rowFilters = [
      {
        name: rowDimensionNames[0],
        in: rowDimensionAxes[rowDimensionNames[0]].slice(0, yLimit),
      },
    ];
  }

  const colFilters = colDimensionNames.map((colDimensionName) => {
    return {
      name: colDimensionName,
      in: colDimensionAxes[colDimensionName].slice(0, xLimit),
    };
  });

  const filters = {
    include: [...colFilters, ...rowFilters],
    exclude: [],
  };

  return filters;
}

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  dimensionNames: string[],
  headers,
  leafData
) {
  if (!dimensionNames.length || !headers) return leafData;

  const colValuesIndexMaps = dimensionNames.map((colDimension) =>
    createIndexMap(headers[colDimension])
  );

  const levels = dimensionNames.length;
  // Recursive function to create nested columns
  function createNestedColumns(level: number, colValuePair) {
    if (level === levels) {
      const accessors = getAccessorForCell(
        dimensionNames,
        colValuesIndexMaps,
        leafData.length,
        colValuePair
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
  config,
  columnDimensionAxes: string[] = []
) {
  const IsNested = true;

  // TODO: Simplify function calls
  const measures = getMeasuresInPivotColumns(config.pivot, config.allMeasures);
  const rowDimensions = getDimensionsInPivotRow(
    config.pivot,
    config.allDimensions
  );
  const colDimensions = getDimensionsInPivotColumns(
    config.pivot,
    config.allDimensions
  );

  let rowDimensionsForColumnDef = rowDimensions;
  let nestedLabel;
  if (IsNested) {
    rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
    nestedLabel = rowDimensions.map((d) => d.label || d.name).join(" > ");
  }
  const rowDefinitions = rowDimensionsForColumnDef.map((d) => {
    return {
      accessorKey: d.name,
      header: nestedLabel ? nestedLabel : d.label || d.name,
      cell: ({ row, getValue }) =>
        cellComponent(PivotExpandableCell, {
          value: getValue(),
          row,
        }),
    };
  });

  const leafColumns = measures.map((m) => {
    return {
      accessorKey: m.name,
      header: m.label || m.name,
      cell: (info) => info.getValue(),
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    colDimensions.map((d) => d.column) as string[],
    columnDimensionAxes,
    leafColumns
  );

  return [...rowDefinitions, ...groupedColDef];
}
