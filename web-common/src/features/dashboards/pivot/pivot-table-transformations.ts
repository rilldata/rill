import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import { createIndexMap, getAccessorForCell } from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";

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
  formatters: Map<
    string,
    (value: string | number | null | undefined) => string | null | undefined
  >,
  isExpanded = false,
) {
  const colDimensionNames = config.colDimensionNames;
  const rowPage = config.pivot.rowPage;
  const rowOffset = isExpanded ? 0 : (rowPage - 1) * NUM_ROWS_PER_PAGE;

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
          const measureName = config.measureNames[i];
          let formatter = formatters.get(measureName);

          console.log({ measureName });

          if (!formatter) {
            if (measureName.endsWith("percent")) {
              console.log("percentage formatter");
              formatter = percentageFormatter;
            } else if (measureName.endsWith("delta")) {
              console.log("delta");
              formatter = formatters.get(
                measureName.replace("__comparison_delta", ""),
              );
            }
          }

          const value = cell[measureName] as string | number;

          row[accessor] = value;
          row[accessor + "__f"] = formatter ? formatter(value) : undefined;
        });
      }
    } else {
      // In absence of any anchor dimension, the first row is the only row

      if (!tableData.length) {
        tableData = [{}];
      }
      const row = tableData[0];

      accessors.forEach((accessor, i) => {
        const measureName = config.measureNames[i];
        let formatter = formatters.get(measureName);

        if (!formatter) {
          if (measureName.endsWith("percent")) {
            formatter = percentageFormatter;
          } else if (measureName.endsWith("delta")) {
            formatter = formatters.get(
              measureName.replace("__comparison_delta", ""),
            );
          }
        }

        const value = cell[measureName] as string | number;

        row[accessor] = value;

        row[accessor + "__f"] = formatter ? formatter(value) : undefined;
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

  const formatters = config.allMeasures.reduce((map, m) => {
    map.set(m.name ?? "", createMeasureValueFormatter(m));
    return map;
  }, new Map<string, ReturnType<typeof createMeasureValueFormatter>>());

  let totalsRow: PivotDataRow = {};
  if (measureNames.length) {
    const totalsRowTable = reduceTableCellDataIntoRows(
      config,
      "",
      [],
      columnDimensionAxes || {},
      [],
      [],
      formatters,
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

  const formatters = config.allMeasures.reduce((map, m) => {
    map.set(m.name ?? "", createMeasureValueFormatter(m));
    return map;
  }, new Map<string, ReturnType<typeof createMeasureValueFormatter>>());

  let totalsRow: PivotDataRow = {};
  if (measureNames.length) {
    const totalsRowTable = reduceTableCellDataIntoRows(
      config,
      "",
      [],
      columnDimensionAxes || {},
      [],
      totalsRowData || [],
      formatters,
    );

    totalsRow = totalsRowTable[0] || {};

    globalTotalsData.forEach((total) => {
      totalsRow = { ...total, ...totalsRow };
    });

    if (anchorDimensionName && !config.isFlat) {
      totalsRow[anchorDimensionName] = "Total";
    }
  }

  return totalsRow;
}

export function mergeRowTotalsInOrder(
  rowValues: string[],
  sortedRowTotals: V1MetricsViewAggregationResponseDataItem[],
  unsortedRowValues: string[],
  unsortedRowTotals: V1MetricsViewAggregationResponseDataItem[],
): V1MetricsViewAggregationResponseDataItem[] {
  if (unsortedRowValues.length === 0) {
    return sortedRowTotals;
  }

  const NOT_AVAILABLE = Symbol("NOT_AVAILABLE");
  const unsortedRowValuesMap = createIndexMap(unsortedRowValues);

  const orderedRowTotals = rowValues
    .map((rowValue) => {
      const unsortedRowIndex = unsortedRowValuesMap.get(rowValue);
      if (unsortedRowIndex === undefined) {
        /**
         * Exclude missing values when sorting by deltas to ensure only dimension values
         * present in both time ranges are returned.
         *
         * This prevents discrepancies between the unsorted and sorted table row sets.
         */
        return NOT_AVAILABLE;
      }
      return unsortedRowTotals[unsortedRowIndex];
    })
    .filter((rowTotal) => rowTotal !== NOT_AVAILABLE);

  return orderedRowTotals;
}

export function percentageFormatter(value: number | null | undefined) {
  if (value === null || value === undefined) return undefined;
  if (value === 0) return "0.0%";

  if (value < 0.005 && value > -0.005) return "~0%";
  console.log({ value });
  return `${value < 0 ? "-" : ""}${(Math.abs(value) * 100).toFixed(2)}%`;

  // return `${value.toFixed(2)}%`;
}
