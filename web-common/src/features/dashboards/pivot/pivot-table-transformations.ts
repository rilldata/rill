import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import { createIndexMap, getAccessorForCell } from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

/**
 * During the phase when queries are still being resolved, we don't have enough
 * information about the values present in an expanded group
 * For now fill it with empty values if there are more than one row dimensions
 */
export function prepareNestedPivotData(
  data: PivotDataRow[],
  dimensions: string[],
  i = 1,
) {
  if (dimensions.slice(i).length > 0) {
    data.forEach((row) => {
      if (!row.subRows) {
        row.subRows = [{ [dimensions[0]]: "LOADING_CELL" }];
      }

      prepareNestedPivotData(row.subRows, dimensions, i + 1);
    });
  }
}

/***
 *
 * The table expects pivoted data for a single row to be mapped to a single object
 * This functions reduces the API response from
 *
 * +-------------------------------+    +------------------------------------------------+
 * | {                             |    | {                                              |
 * |  "row_dimension: value,       |    |   "row_dimension": value,                      |
 * |  "column_dimension: value_1,  | => |   "column_dimension_value_1": measure_value_1, |
 * |  "measure": measure_value_1   |    |   "column_dimension_value_2": measure_value_2  |
 * | },                            |    | }                                              |
 * | {                             |    +------------------------------------------------+
 * |  "row_dimension: value,       |    Transformed to a Tanstack table readable format
 * |  "column_dimension: value_2,  |
 * |  "measure": measure_value_2   |
 * | }                             |
 * +-------------------------------+
 */
export function reduceTableCellDataIntoRows(
  config: PivotDataStoreConfig,
  anchorDimensionName: string,
  anchorDimensionRowValues: string[],
  columnDimensionAxes: Record<string, string[]>,
  tableData: PivotDataRow[],
  cellData: V1MetricsViewAggregationResponseDataItem[],
) {
  const colDimensionNames = config.colDimensionNames;
  const rowPage = config.pivot.rowPage;
  const rowOffset = (rowPage - 1) * NUM_ROWS_PER_PAGE;

  /**
   * Create a map of row dimension values to their index in the row dimension axes.
   * This way we can apply the sort order on the row dimension axes and sort the cell
   * data using the dimension values. For O(n) pass on the cells (n = number of cells),
   * we can reduce them optimally into row objects.
   */
  const rowDimensionIndexMap = createIndexMap(anchorDimensionRowValues);

  const colValuesIndexMaps = colDimensionNames.map((colDimensionName) =>
    createIndexMap(columnDimensionAxes[colDimensionName]),
  );

  cellData?.forEach((cell) => {
    const accessors = getAccessorForCell(
      colDimensionNames,
      config.time.timeDimension,
      colValuesIndexMaps,
      config.measureNames.length,
      cell,
    );

    if (anchorDimensionName) {
      const rowDimensionValue = cell[anchorDimensionName] as string;
      const rowIndex = rowDimensionIndexMap.get(rowDimensionValue);
      if (rowIndex === undefined) {
        return;
      }
      const row = tableData[rowOffset + rowIndex];

      if (row) {
        accessors.forEach((accessor, i) => {
          row[accessor] = cell[config.measureNames[i]] as string | number;
        });
      }
    } else {
      // In absence of any anchor dimension, the first row is the only row

      if (!tableData.length) {
        tableData = [{}];
      }
      const row = tableData[0];

      accessors.forEach((accessor, i) => {
        row[accessor] = cell[config.measureNames[i]] as string | number;
      });

      return;
    }
  });

  return tableData;
}

export function getTotalsRowSkeleton(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
) {
  const { rowDimensionNames, measureNames } = config;
  const anchorDimensionName = rowDimensionNames[0];

  let totalsRow: PivotDataRow = {};
  if (measureNames.length) {
    const totalsRowTable = reduceTableCellDataIntoRows(
      config,
      "",
      [],
      columnDimensionAxes || {},
      [],
      [],
    );

    totalsRow = totalsRowTable[0] || {};

    if (anchorDimensionName) {
      totalsRow[anchorDimensionName] = "Total";
    }
  }

  return totalsRow;
}

export function getTotalsRow(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
  totalsRowData: V1MetricsViewAggregationResponseDataItem[] = [],
  globalTotalsData: V1MetricsViewAggregationResponseDataItem[] = [],
) {
  const { rowDimensionNames, measureNames } = config;
  const anchorDimensionName = rowDimensionNames[0];

  let totalsRow: PivotDataRow = {};
  if (measureNames.length) {
    const totalsRowTable = reduceTableCellDataIntoRows(
      config,
      "",
      [],
      columnDimensionAxes || {},
      [],
      totalsRowData || [],
    );

    totalsRow = totalsRowTable[0] || {};

    globalTotalsData.forEach((total) => {
      totalsRow = { ...total, ...totalsRow };
    });

    if (anchorDimensionName) {
      totalsRow[anchorDimensionName] = "Total";
    }
  }

  return totalsRow;
}

export function mergeRowTotals(
  rowValues: string[],
  sortedRowTotals: V1MetricsViewAggregationResponseDataItem[],
  unsortedRowValues: string[],
  unsortedRowTotals: V1MetricsViewAggregationResponseDataItem[],
): V1MetricsViewAggregationResponseDataItem[] {
  if (unsortedRowValues.length === 0) {
    return sortedRowTotals;
  }

  const unsortedRowValuesMap = createIndexMap(unsortedRowValues);

  const orderedRowTotals = rowValues.map((rowValue) => {
    const unsortedRowIndex = unsortedRowValuesMap.get(rowValue);
    if (unsortedRowIndex === undefined) {
      console.error("Row value not found in unsorted row values", rowValue);
    }
    const rowTotal = unsortedRowTotals[unsortedRowIndex as number];
    return rowTotal;
  });

  return orderedRowTotals;
}
