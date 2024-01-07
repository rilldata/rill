/**
 * Handles transformations for expanded rows in a nested pivot table
 */

import { createPivotAggregationRowQuery } from "./pivot-data-store";
import type { ExpandedState } from "@tanstack/svelte-table";
import { derived, writable } from "svelte/store";

function getExpandedValuesFromNestedArray(
  dataArray,
  anchorDimension: string,
  expanded: ExpandedState
): Record<string, string[]> {
  const values = {};

  for (const key in expanded as Record<string, boolean>) {
    if (expanded[key]) {
      // Split the key into indices
      const indices = key.split(".").map((index) => parseInt(index, 10));

      // Retrieve the value from the nested array
      let currentValue = dataArray;
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

export function queryExpandedRowMeasureValues(
  ctx,
  data,
  measures: string[],
  allDimensions: string[],
  expanded: ExpandedState
) {
  if (!data || Object.keys(expanded).length == 0) return writable(null);
  const values = getExpandedValuesFromNestedArray(
    data,
    allDimensions[0],
    expanded
  );

  return derived(
    Object.keys(values)?.map((expandIndex) => {
      const dimensions = [allDimensions[values[expandIndex].length]];
      // TODO: handle for already existing filters
      const includeFilters = values[expandIndex].map((value, index) => {
        return {
          name: allDimensions[index],
          in: [value],
        };
      });

      const filters = {
        include: includeFilters,
        exclude: [],
      };
      return derived(
        [
          writable(expandIndex),
          createPivotAggregationRowQuery(ctx, measures, dimensions, filters),
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
