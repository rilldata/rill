import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
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
      const row = tableData[rowIndex];

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

const PivotMeasureRegex = /(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\+\d{2})_(.*)/;
export function transformPivotToRows(
  aggregationRows: V1MetricsViewAggregationResponseDataItem[],
  timeDimension: string,
  dimensions: string[],
  measures: string[],
): V1MetricsViewAggregationResponseDataItem[] {
  const dimensionMap = getMapFromArray(dimensions, (d) => d);
  const allRows: V1MetricsViewAggregationResponseDataItem[] = [];

  for (const aggregationRow of aggregationRows) {
    const dimValuesRow: V1MetricsViewAggregationResponseDataItem = {};
    dimensions.forEach((d) => (dimValuesRow[d] = aggregationRow[d]));

    const rowByTime = new Map<
      string,
      V1MetricsViewAggregationResponseDataItem
    >();

    for (const k in aggregationRow) {
      if (dimensionMap.has(k)) continue;
      const matches = PivotMeasureRegex.exec(k);
      if (matches?.length !== 3) continue;

      const [, timestamp, measure] = matches;

      let row: V1MetricsViewAggregationResponseDataItem;
      if (rowByTime.has(timestamp)) {
        row = rowByTime.get(
          timestamp,
        ) as V1MetricsViewAggregationResponseDataItem;
      } else {
        row = {
          ...dimValuesRow,
          [timeDimension]: timestamp,
        };
        rowByTime.set(timestamp, row);
      }

      row[measure] = aggregationRow[k];
    }

    allRows.push(
      ...Array.from(rowByTime.values()).filter((r) =>
        measures.every((m) => r[m] !== null),
      ),
    );
  }

  allRows.sort((a, b) => (a[timeDimension] <= b[timeDimension] ? 1 : -1));
  return allRows;
}
