import {
  createQueryServiceMetricsViewAggregation,
  V1MetricsViewFilter,
  type V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, Readable, writable } from "svelte/store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import { queryExpandedRowMeasureValues } from "./pivot-expansion";
import {
  getDimensionsInPivotColumns,
  getDimensionsInPivotRow,
  getFilterForPivotTable,
  getMeasuresInPivotColumns,
  prepareExpandedPivotData,
} from "./pivot-utils";
import {
  createTableWithAxes,
  reduceTableCellDataIntoRows,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-table-transformations";

/**
 * Extract out config relevant to pivot from dashboard and meta store
 */
function getPivotConfig(ctx: StateManagers) {
  return derived(
    [useMetaQuery(ctx), ctx.dashboardStore],
    ([metricsView, dashboardStore]) => {
      const { rows, columns } = dashboardStore.pivot;

      if (
        (rows.length == 0 && columns.length == 0) ||
        !metricsView.data?.measures ||
        !metricsView.data?.dimensions
      ) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          filters: dashboardStore.filters,
          pivot: dashboardStore.pivot,
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
      const rowDimensionNames = dimensions.map((d) => d.column) as string[];
      const colDimensionNames = columnDimensons.map(
        (d) => d.column
      ) as string[];
      return {
        measureNames,
        rowDimensionNames,
        colDimensionNames,
        allMeasures: metricsView.data?.measures,
        allDimensions: metricsView.data?.dimensions,
        filters: dashboardStore.filters,
        pivot: dashboardStore.pivot,
      };
    }
  );
}

/**
 * Wrapper function for Aggregate Query API
 */
export function createPivotAggregationRowQuery(
  ctx: StateManagers,
  measures: string[],
  dimensions: string[],
  filters: V1MetricsViewFilter,
  sort: V1MetricsViewAggregationSort[] = [],
  limit = "100",
  offset = "0"
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  // Todo: Handle sorting in table
  if (!sort.length) {
    sort = [
      {
        desc: false,
        name: measures[0] || dimensions[0],
      },
    ];
  }
  return derived(
    [ctx.runtime, ctx.metricsViewName, useTimeControlStore(ctx)],
    ([runtime, metricViewName, timeControls], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          dimensions: dimensions.map((dimension) => ({ name: dimension })),
          filter: filters,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          sort,
          limit,
          offset,
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
export function getAxisForDimensions(ctx, dimensions, filters) {
  if (!dimensions.length) return writable(null);
  return derived(
    dimensions.map((dimension) =>
      createPivotAggregationRowQuery(ctx, [], [dimension], filters)
    ),
    (data: Array<any>) => {
      const axesMap = {};

      // Wait for all data to populate
      if (data.some((d) => d?.isFetching)) return { isFetching: true };

      data.forEach((d, i: number) => {
        axesMap[dimensions[i]] = d?.data?.data?.map(
          (dimValue) => dimValue[dimensions[i]]
        );
      });

      return { isFetching: false, data: axesMap };
    }
  );
}

/**
 * Main store for pivot table data
 *
 * At a high-level, we make the following queries in the order below:
 *
 * Input pivot config
 *     |
 *     |  (Axes)
 *     v
 * Create table headers by querying axes values for each dimension
 *     |
 *     |  (Cell Data)
 *     v
 * For the visible axes values, query the data for each cell, totals and subtotals
 *     |
 *     |  (Expanded)
 *     v
 * For each expanded row, query the data for each cell, totals and subtotals
 *     |
 *     |  (Assemble)
 *     v
 * Table data and column definitions
 */
function createPivotDataStore(ctx: StateManagers): PivotDataStore {
  /**
   * Derive a store using pivot config
   */
  return derived(getPivotConfig(ctx), (config, set) => {
    const { measureNames, rowDimensionNames, colDimensionNames, pivot } =
      config;

    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
      colDimensionNames,
      config.filters
    );

    const rowDimensionAxisQuery = getAxisForDimensions(
      ctx,
      rowDimensionNames,
      config.filters
    );

    /**
     * Derive a store from axes queries
     */
    return derived(
      [columnDimensionAxesQuery, rowDimensionAxisQuery],
      ([columnDimensionAxes, rowDimensionAxes], axesSet) => {
        if (columnDimensionAxes?.isFetching || rowDimensionAxes?.isFetching) {
          return { isFetching: true };
        }

        const skeletonTable = createTableWithAxes(
          config,
          columnDimensionAxes?.data,
          rowDimensionAxes?.data
        );

        const columnDef = skeletonTable.columnDef;
        let tableData = skeletonTable.data;

        const rowDimensionName = rowDimensionNames.slice(0, 1);
        const allDimensions = colDimensionNames.concat(rowDimensionName);
        const filterForInitialTable = getFilterForPivotTable(
          config,
          columnDimensionAxes?.data,
          rowDimensionAxes?.data
        );

        const sortBy = [
          {
            desc: false,
            name: rowDimensionName[0],
          },
        ];
        const initialTableCellQuery = createPivotAggregationRowQuery(
          ctx,
          measureNames,
          allDimensions,
          filterForInitialTable,
          sortBy,
          "10000"
        );

        /**
         * Derive a using initial table view
         */
        return derived(
          [initialTableCellQuery],
          ([initialTableCellData], set2) => {
            // Wait for data
            if (initialTableCellData.isFetching)
              return { isFetching: false, data: tableData, columnDef };
            if (initialTableCellData.error)
              return { isFetching: false, data: [] };

            let cellData = initialTableCellData.data?.data;

            console.log("initialTableCellDataQueryResult", cellData);

            tableData = reduceTableCellDataIntoRows(
              config,
              columnDimensionAxes?.data,
              rowDimensionAxes?.data,
              tableData,
              cellData
            );

            console.log("tableData2", tableData);

            /**
             * Derive a store based on expanded rows
             */
            return derived(
              queryExpandedRowMeasureValues(
                ctx,
                cellData,
                measureNames,
                rowDimensionNames,
                pivot.expanded
              ),
              (expandedRowMeasureValues) => {
                prepareExpandedPivotData(
                  cellData,
                  rowDimensionNames,
                  pivot.expanded
                );

                if (expandedRowMeasureValues?.length) {
                  cellData = addExpandedDataToPivot(
                    cellData,
                    rowDimensionNames,
                    expandedRowMeasureValues
                  );
                }
                return { isFetching: false, data: tableData, columnDef };
              }
            ).subscribe(set2);
          }
        ).subscribe(axesSet);
      }
    ).subscribe(set);
  });
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
