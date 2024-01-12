import {
  createQueryServiceMetricsViewAggregation,
  V1MetricsViewFilter,
  type V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
  V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, Readable, writable } from "svelte/store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import {
  addExpandedDataToPivot,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import {
  getColumnDefForPivot,
  getDimensionsInPivotColumns,
  getDimensionsInPivotRow,
  getFilterForPivotTable,
  getMeasuresInPivotColumns,
} from "./pivot-utils";
import {
  createTableWithAxes,
  reduceTableCellDataIntoRows,
  prepareNestedPivotData,
} from "./pivot-table-transformations";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

/**
 * Extract out config relevant to pivot from dashboard and meta store
 */
function getPivotConfig(ctx: StateManagers): Readable<PivotDataStoreConfig> {
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
        metricsView.data?.measures,
      );
      const dimensions = getDimensionsInPivotRow(
        dashboardStore.pivot,
        metricsView.data?.dimensions,
      );

      const columnDimensons = getDimensionsInPivotColumns(
        dashboardStore.pivot,
        metricsView.data?.dimensions,
      );

      const measureNames = measures.map((m) => m.name) as string[];
      const rowDimensionNames = dimensions.map((d) => d.column) as string[];
      const colDimensionNames = columnDimensons.map(
        (d) => d.column,
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
    },
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
  offset = "0",
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
        },
      ).subscribe(set),
  );
}

/**
 * Returns a query for cell data for the initial table.
 * TODO: Add description for sorting methodolgy
 */
export function createTableCellQuery(
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  anchorDimension: string,
  columnDimensionAxesData: Record<string, string[]> | undefined,
  rowDimensionAxesData: Record<string, string[]> | undefined,
) {
  let allDimensions = config.colDimensionNames;
  if (anchorDimension) {
    allDimensions = config.colDimensionNames.concat([anchorDimension]);
  }

  const filterForInitialTable = getFilterForPivotTable(
    config,
    columnDimensionAxesData,
    rowDimensionAxesData,
    true,
  );

  const sortBy = [
    {
      desc: false,
      name: anchorDimension || config.measureNames[0],
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    config.measureNames,
    allDimensions,
    filterForInitialTable,
    sortBy,
    "10000",
  );
}

/***
 * Get a list of axis values for a given list of dimension values and filters
 */
export function getAxisForDimensions(
  ctx: StateManagers,
  dimensions: string[],
  filters: V1MetricsViewFilter,
  sortBy: V1MetricsViewAggregationSort[] = [],
) {
  if (!dimensions.length) return writable(null);

  // FIXME: If sorting by measure, add that to measure list
  let measures: string[] = [];
  if (sortBy.length) {
    const sortMeasure = sortBy[0].name as string;
    if (!dimensions.includes(sortMeasure)) {
      measures = [sortMeasure];
    }
  }

  return derived(
    dimensions.map((dimension) =>
      createPivotAggregationRowQuery(
        ctx,
        measures,
        [dimension],
        filters,
        sortBy,
      ),
    ),
    (data) => {
      const axesMap: Record<string, string[]> = {};

      // Wait for all data to populate
      if (data.some((d) => d?.isFetching)) return { isFetching: true };

      data.forEach((d, i: number) => {
        const dimensionName = dimensions[i];
        axesMap[dimensionName] = (d?.data?.data || [])?.map(
          (dimValue) => dimValue[dimensionName] as string,
        );
      });

      if (Object.values(axesMap).some((d) => !d)) return { isFetching: true };

      return {
        isFetching: false,
        data: axesMap,
      };
    },
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
    const { rowDimensionNames, colDimensionNames, measureNames } = config;

    if (!rowDimensionNames.length && !measureNames.length) {
      return set({ isFetching: false, data: [] });
    }

    const sortPivotBy = config.pivot.sorting.map((sort) => ({
      name: sort.id,
      desc: sort.desc,
    }));

    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
      colDimensionNames,
      config.filters,
    );

    const rowDimensionAxisQuery = getAxisForDimensions(
      ctx,
      rowDimensionNames,
      config.filters,
      sortPivotBy,
    );

    /**
     * Derive a store from axes queries
     */
    return derived(
      [columnDimensionAxesQuery, rowDimensionAxisQuery],
      ([columnDimensionAxes, rowDimensionAxes], axesSet) => {
        if (columnDimensionAxes?.isFetching || rowDimensionAxes?.isFetching) {
          return axesSet({ isFetching: true });
        }

        const anchorDimension = rowDimensionNames[0];
        const skeletonTableData = createTableWithAxes(
          anchorDimension,
          rowDimensionAxes?.data?.[anchorDimension],
        );

        const columnDef = getColumnDefForPivot(
          config,
          columnDimensionAxes?.data,
        );

        const initialTableCellQuery = createTableCellQuery(
          ctx,
          config,
          rowDimensionNames[0],
          columnDimensionAxes?.data,
          rowDimensionAxes?.data,
        );

        /**
         * Derive a store from initial table cell data query
         */
        return derived(
          [initialTableCellQuery],
          ([initialTableCellData], cellSet) => {
            // Wait for data
            if (initialTableCellData.isFetching || initialTableCellData.error)
              // return cellSet({
              //   isFetching: false,
              //   data: skeletonTableData,
              //   columnDef,
              // });

              // FIXME: Table does not render properly if below object
              // is set using derived stores set method
              return {
                isFetching: false,
                data: skeletonTableData,
                columnDef,
              };

            const cellData = initialTableCellData.data
              ?.data as V1MetricsViewAggregationResponseDataItem[];

            const tableDataWithCells = reduceTableCellDataIntoRows(
              config,
              anchorDimension,
              rowDimensionAxes?.data?.[anchorDimension] || [],
              columnDimensionAxes?.data || {},
              skeletonTableData,
              cellData,
            );

            const expandedSubTableCellQuery = queryExpandedRowMeasureValues(
              ctx,
              config,
              tableDataWithCells,
              columnDimensionAxes?.data,
            );
            /**
             * Derive a store based on expanded rows
             */
            return derived(
              expandedSubTableCellQuery,
              (expandedRowMeasureValues) => {
                prepareNestedPivotData(tableDataWithCells, rowDimensionNames);
                let tableDataExpanded: PivotDataRow[] = tableDataWithCells;
                if (expandedRowMeasureValues?.length) {
                  tableDataExpanded = addExpandedDataToPivot(
                    config,
                    tableDataWithCells,
                    rowDimensionNames,
                    columnDimensionAxes?.data || {},
                    expandedRowMeasureValues,
                  );
                }
                return {
                  isFetching: false,
                  data: tableDataExpanded,
                  columnDef,
                };
              },
            ).subscribe(cellSet);
          },
        ).subscribe(axesSet);
      },
    ).subscribe(set);
  });
}

interface PivotDataState {
  isFetching: boolean;
  data?: PivotDataRow[];
  columnDef?: Array<unknown>;
}

export type PivotDataStore = Readable<PivotDataState>;

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const usePivotDataStore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => createPivotDataStore(ctx),
);
