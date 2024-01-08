import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { createPivotAggregationRowQuery } from "./pivot-data-store";
import type { ExpandedState } from "@tanstack/svelte-table";
import { derived, writable } from "svelte/store";
import type { PivotDataStoreConfig } from "@rilldata/web-common/features/dashboards/pivot/types";
import { getFilterForPivotTable } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";

/**
 * Extracts and organizes dimension names from a nested array structure
 * based on a specified anchor dimension and an expanded state.
 *
 * This function iterates over each key in the `expanded` object, which
 * indicates whether a particular path in the nested array is expanded.
 * For each expanded path, it navigates through the table data
 * following the path defined by the key (split into indices) and extracts
 * the dimension values at each level.
 *
 */
function getExpandedValuesFromNestedArray(
  tableData,
  anchorDimension: string,
  expanded: ExpandedState
): Record<string, string[]> {
  const values = {};

  for (const key in expanded as Record<string, boolean>) {
    if (expanded[key]) {
      // Split the key into indices
      const indices = key.split(".").map((index) => parseInt(index, 10));

      // Retrieve the value from the nested array
      let currentValue = tableData;
      const dimensionNames: string[] = [];
      for (const index of indices) {
        if (!currentValue?.[index]) break;
        dimensionNames.push(currentValue[index]?.[anchorDimension]);
        currentValue = currentValue[index]?.subRows;
      }

      // Add the value to the result array
      values[key] = dimensionNames;
    }
  }

  return values;
}

/**
 * Returns a query for cell data for a sub table
 */
export function createSubTableCellQuery(
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  anchorDimension: string,
  columnDimensionAxesData,
  rowNestFilters
) {
  const allDimensions = config.colDimensionNames.concat([anchorDimension]);

  const filterForSubTable = getFilterForPivotTable(
    config,
    columnDimensionAxesData
  );

  const includeFilters = filterForSubTable.include.concat(rowNestFilters);
  const filters = {
    include: includeFilters,
    exclude: [],
  };

  const sortBy = [
    {
      desc: false,
      name: anchorDimension,
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    config.measureNames,
    allDimensions,
    filters,
    sortBy,
    "10000"
  );
}

export function queryExpandedRowMeasureValues(
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  tableData,
  columnDimensionAxesData
) {
  const { rowDimensionNames } = config;
  const expanded = config.pivot.expanded;
  if (!tableData || Object.keys(expanded).length == 0) return writable(null);
  const values = getExpandedValuesFromNestedArray(
    tableData,
    rowDimensionNames[0],
    expanded
  );

  return derived(
    Object.keys(values)?.map((expandIndex) => {
      const anchorDimension = rowDimensionNames[values[expandIndex].length];
      // TODO: handle for already existing filters
      const rowNestFilters = values[expandIndex].map((value, index) => {
        return {
          name: rowDimensionNames[index],
          in: [value],
        };
      });

      return derived(
        [
          writable(expandIndex),
          createSubTableCellQuery(
            ctx,
            config,
            anchorDimension,
            columnDimensionAxesData,
            rowNestFilters
          ),
        ],
        ([expandIndex, query]) => {
          return {
            isFetching: query?.isFetching,
            expandIndex,
            data: query?.data?.data,
          };
        }
      );
    }),
    (combos) => {
      return combos;
    }
  );
}

export function addExpandedDataToPivot(
  data,
  dimensions,
  expandedRowMeasureValues
) {
  const pivotData = data;
  const levels = dimensions.length;

  expandedRowMeasureValues.forEach((expandedRowData) => {
    const indices = expandedRowData.expandIndex
      .split(".")
      .map((index) => parseInt(index, 10));

    let parent = pivotData; // Keep a reference to the parent array
    let lastIdx = 0; // Keep track of the last index

    // Traverse the data array to the right position
    for (let i = 0; i < indices.length; i++) {
      if (!parent[indices[i]]) break;
      if (i < indices.length - 1) {
        parent = parent[indices[i]].subRows;
      }
      lastIdx = indices[i];
    }

    // Update the specific array at the position
    if (parent[lastIdx] && parent[lastIdx].subRows) {
      if (!expandedRowData?.data?.length) {
        parent[lastIdx].subRows = [{ [dimensions[0]]: "LOADING_CELL" }];
      } else {
        parent[lastIdx].subRows = expandedRowData?.data.map((row) => {
          const newRow = {
            ...row,
            [dimensions[0]]: row[dimensions[indices.length]],
          };

          if (indices.length < levels - 1) {
            newRow.subRows = [{ [dimensions[0]]: "LOADING_CELL" }];
          }
          return newRow;
        });
      }
    }
  });
  return pivotData;
}
