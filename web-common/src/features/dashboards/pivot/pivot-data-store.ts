import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, Readable, writable } from "svelte/store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import { ExpandedState, flexRender } from "@tanstack/svelte-table";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import {
  getColumnDefForPivot,
  getDimensionsInPivotColumns,
  getDimensionsInPivotRow,
  getMeasuresInPivotColumns,
  prepareExpandedPivotData,
} from "./pivot-utils";

function createPivotAggregationRowQuery(
  ctx: StateManagers,
  measures,
  dimensions,
  filters
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricViewName, dashboardStore, timeControls], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          dimensions: dimensions.map((dimension) => ({ name: dimension })),
          filter: filters,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          sort: [
            {
              desc: dashboardStore.sortDirection === SortDirection.DESCENDING,
              name: measures[0] || dimensions[0],
            },
          ],
          limit: "50",
          offset: "0",
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        }
      ).subscribe(set)
  );
}

/***
 * Get a list of axis values for a given list of dimension values and filters
 */
function getAxisForDimensions(ctx, dimensions, filters) {
  return derived(
    dimensions.map((dimension) =>
      createPivotAggregationRowQuery(ctx, [], [dimension.column], filters)
    ),
    (data) => {
      if (data.some((d) => d.isFetching)) return undefined;
      return data.map((d) => d?.data?.data);
    }
  );
}

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

function queryExpandedRowMeasureValues(
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

function addExpandedDataToPivot(data, dimensions, expandedRowMeasureValues) {
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

function createPivotDataStore(ctx: StateManagers): PivotDataStore {
  const isNested = true;
  return derived(
    [useMetaQuery(ctx), useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsView, timeControls, dashboardStore], set) => {
      const { rows, columns, expanded } = dashboardStore.pivot;

      if (!timeControls.ready || timeControls.isFetching) {
        return { isFetching: true };
      }
      if (rows.length == 0 && columns.length == 0) {
        return { isFetching: false, data: [] };
      }
      if (!metricsView.data?.measures || !metricsView.data?.dimensions) {
        return {
          isFetching: true,
          data: [],
        };
      }

      const measures = getMeasuresInPivotColumns(
        dashboardStore.pivot,
        metricsView.data?.measures
      );
      const dimensions = getDimensionsInPivotRow(
        dashboardStore.pivot,
        metricsView.data?.dimensions
      );

      const columnDimensons = getDimensionsInPivotColumns(
        dashboardStore.pivot,
        metricsView.data?.dimensions
      );

      const measureNames = measures.map((m) => m.name) as string[];
      const dimensionNames = dimensions.map((d) => d.column) as string[];

      let dimensionForInitialView = dimensions;
      let columnDimensionAxesQuery = writable(null);

      if (isNested && dimensions.length > 1) {
        dimensionForInitialView = dimensions.slice(0, 1);

        if (columnDimensons.length) {
          dimensionForInitialView =
            dimensionForInitialView.concat(columnDimensons);

          columnDimensionAxesQuery = getAxisForDimensions(
            ctx,
            columnDimensons,
            dashboardStore.filters
          );
        }
      }

      const initialTableView = createPivotAggregationRowQuery(
        ctx,
        measures.map((m) => m.name),
        dimensionForInitialView.map((d) => d.column),
        dashboardStore.filters
      );

      return derived(
        [initialTableView, columnDimensionAxesQuery],
        ([initialTable, columnDimensionData], set2) => {
          // Wait for data
          if (initialTable.isFetching) return { isFetching: true };
          if (initialTable.error) return { isFetching: false, data: [] };

          let data = initialTable.data?.data;

          if (columnDimensons.length && !columnDimensionData) {
            return { isFetching: true };
          }

          console.log("columnDimensionData", columnDimensionData);

          return derived(
            queryExpandedRowMeasureValues(
              ctx,
              data,
              measureNames,
              dimensionNames,
              expanded
            ),
            (expandedRowMeasureValues) => {
              prepareExpandedPivotData(data, dimensionNames, expanded);

              if (expandedRowMeasureValues?.length) {
                data = addExpandedDataToPivot(
                  data,
                  dimensionNames,
                  expandedRowMeasureValues
                );
              }
              return {
                isFetching: false,
                data,
                columnDef: getColumnDefForPivot(
                  dashboardStore?.pivot,
                  metricsView?.data,
                  columnDimensionData || []
                ),
              };
            }
          ).subscribe(set2);
        }
      ).subscribe(set);
    }
  );
}

interface PivotDataState {
  isFetching: boolean;
  data?: Array<unknown>;
  columnDef?: Array<unknown>;
}

export type PivotDataStore = Readable<PivotDataState>;

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const usePivotDataStore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => createPivotDataStore(ctx)
);
