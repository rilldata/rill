import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import type { PivotState } from "./types";

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

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  levels: number,
  dimensions: MetricsViewSpecDimensionV2[],
  headers,
  leafData
) {
  if (!dimensions.length || !headers?.length) return leafData;
  // Recursive function to create nested columns
  function createNestedColumns(level: number, headerValues) {
    console.log(headerValues, dimensions, level);
    if (level === levels) {
      // Base case: return leaf columns
      return leafData;
    }

    // Recursive case: create nested headers
    return headerValues.map((header) => ({
      Header: header?.[dimensions[level]?.column],
      columns: createNestedColumns(level + 1, headers[level + 1]),
    }));
  }

  // Start the recursion
  return createNestedColumns(0, headers[0]);
}

export function getColumnDefForPivot(
  pivot: PivotState,
  metricSpec: V1MetricsViewSpec,
  columnDimensionAxes = []
) {
  const IsNested = true;
  const measureCols = getMeasuresInPivotColumns(pivot, metricSpec.measures);
  const dimensionRows = getDimensionsInPivotRow(pivot, metricSpec.dimensions);
  const dimensionCols = getDimensionsInPivotColumns(
    pivot,
    metricSpec.dimensions
  );

  let rowDimensionsForColumnDef = dimensionRows;
  let nestedLabel;
  if (IsNested) {
    rowDimensionsForColumnDef = dimensionRows.slice(0, 1);
    nestedLabel = dimensionRows.map((d) => d.label || d.name).join(" > ");
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

  const leafColumns = measureCols.map((m) => {
    return {
      accessorKey: m.name,
      header: m.label || m.name,
      cell: (info) => info.getValue(),
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    dimensionCols.length,
    dimensionCols,
    columnDimensionAxes,
    leafColumns
  );

  return [...rowDefinitions, ...groupedColDef];
}
