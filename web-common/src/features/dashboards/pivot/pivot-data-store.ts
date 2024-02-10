import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
import {
  V1TimeGrain,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query/build/lib/types";
import type { ColumnDef } from "@tanstack/svelte-table";
import { Readable, derived, readable } from "svelte/store";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  addExpandedDataToPivot,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import { sliceColumnAxesDataForDef } from "./pivot-infinite-scroll";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
} from "./pivot-queries";
import {
  prepareNestedPivotData,
  reduceTableCellDataIntoRows,
} from "./pivot-table-transformations";
import {
  getFilterForPivotTable,
  getPivotConfigKey,
  getSortForAccessor,
  getTotalColumnCount,
  reconcileMissingDimensionValues,
} from "./pivot-utils";
import {
  PivotChipType,
  type PivotDataRow,
  type PivotDataStore,
  type PivotDataStoreConfig,
  type PivotTimeConfig,
} from "./types";

/**
 * Extract out config relevant to pivot from dashboard and meta store
 */
function getPivotConfig(ctx: StateManagers): Readable<PivotDataStoreConfig> {
  return derived(
    [useMetricsView(ctx), ctx.dashboardStore, useTimeControlStore(ctx)],
    ([metricsView, dashboardStore, timeControls]) => {
      let interval: AvailableTimeGrain = "TIME_GRAIN_HOUR";
      const existingTimeGrain = timeControls?.selectedTimeRange?.interval;

      if (existingTimeGrain) {
        if (
          existingTimeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
          existingTimeGrain !== V1TimeGrain.TIME_GRAIN_MILLISECOND &&
          existingTimeGrain !== V1TimeGrain.TIME_GRAIN_SECOND
        )
          interval = existingTimeGrain;
      }

      const time: PivotTimeConfig = {
        timeStart: timeControls.timeStart,
        timeEnd: timeControls.timeEnd,
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView?.data?.timeDimension || "",
        interval,
      };

      if (
        !metricsView.data?.measures ||
        !metricsView.data?.dimensions ||
        !timeControls.ready
      ) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          whereFilter: dashboardStore.whereFilter,
          pivot: dashboardStore.pivot,
          time,
        };
      }

      const measureNames = dashboardStore.pivot.columns.measure.map(
        (m) => m.id,
      );

      // This is temporary until we have a better way to handle time grains
      const rowDimensionNames = dashboardStore.pivot.rows.dimension.map((d) => {
        if (d.type === PivotChipType.Time) {
          time.interval = d.id as AvailableTimeGrain;
          return time.timeDimension;
        }

        return d.id;
      });

      // This is temporary until we have a better way to handle time grains
      const colDimensionNames = dashboardStore.pivot.columns.dimension.map(
        (d) => {
          if (d.type === PivotChipType.Time) {
            time.interval = d.id as AvailableTimeGrain;
            return time.timeDimension;
          }

          return d.id;
        },
      );

      return {
        measureNames,
        rowDimensionNames,
        colDimensionNames,
        allMeasures: metricsView.data?.measures,
        allDimensions: metricsView.data?.dimensions,
        whereFilter: dashboardStore.whereFilter,
        pivot: dashboardStore.pivot,
        time,
      };
    },
  );
}

/**
 * Returns a query for cell data for the initial table.
 * The table cell is sorted by the anchor dimension irrespective
 * of the sort config. The dimension axes values are sorted using
 * the config and values from this query are used to create the
 * table.
 */
export function createTableCellQuery(
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  anchorDimension: string | undefined,
  columnDimensionAxesData: Record<string, string[]> | undefined,
  rowDimensionValues: string[],
) {
  let allDimensions = config.colDimensionNames;
  if (anchorDimension) {
    allDimensions = config.colDimensionNames.concat([anchorDimension]);
  }

  const { time } = config;
  const dimensionBody = allDimensions.map((dimension) => {
    if (dimension === time.timeDimension) {
      return {
        name: dimension,
        timeGrain: time.interval,
        timeZone: time.timeZone,
      };
    } else return { name: dimension };
  });

  const filterForInitialTable = getFilterForPivotTable(
    config,
    columnDimensionAxesData,
    rowDimensionValues,
    true,
    anchorDimension,
  );

  const mergedFilter = mergeFilters(filterForInitialTable, config.whereFilter);

  const sortBy = [
    {
      desc: false,
      name: anchorDimension || config.measureNames[0],
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    config.measureNames,
    dimensionBody,
    mergedFilter,
    sortBy,
    "10000",
  );
}

/**
 * Stores the last pivot data and column def to be used when there is no data
 * to be displayed. This is to avoid the table from flickering when there is no
 * data to be displayed.
 */
let lastPivotData: PivotDataRow[] = [];
let lastPivotColumnDef: ColumnDef<PivotDataRow>[] = [];

/**
 * The expanded table has to iterate over itself to find nested dimension values
 * which are being expanded. Since the expanded values are added in one go, the previously
 * expanded values are not available in the table data. This map stores the expanded table
 * data for each pivot config. This is cleared when the pivot config changes.
 */
let expandedTableMap: Record<string, PivotDataRow[]> = {};

/**
 * Main store for pivot table data
 *
 * At a high-level, we make the following queries in the order below:
 *
 * Input pivot config
 *     |
 *     |  (Column headers)
 *     v
 * Create table headers by querying axes values for each column dimension
 *     |
 *     |  (Row headers and sort order)
 *     v
 * Create skeleton table data by querying axes values for each row dimension
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
  return derived(getPivotConfig(ctx), (config, configSet) => {
    const { rowDimensionNames, colDimensionNames, measureNames } = config;

    if (
      (!rowDimensionNames.length && !measureNames.length) ||
      (colDimensionNames.length && !measureNames.length)
    ) {
      return configSet({
        isFetching: false,
        data: [],
        columnDef: [],
        assembled: false,
        totalColumns: 0,
      });
    }
    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
      config,
      colDimensionNames,
      config.whereFilter,
    );

    return derived(
      columnDimensionAxesQuery,
      (columnDimensionAxes, columnSet) => {
        if (columnDimensionAxes?.isFetching) {
          return columnSet({
            isFetching: true,
            data: [],
            columnDef: [],
            assembled: false,
            totalColumns: 0,
          });
        }
        const anchorDimension = rowDimensionNames[0];

        const { where, sortPivotBy, timeRange } = getSortForAccessor(
          anchorDimension,
          config,
          columnDimensionAxes?.data,
        );

        const totalColumns = getTotalColumnCount(columnDimensionAxes?.data);

        const rowDimensionAxisQuery = getAxisForDimensions(
          ctx,
          config,
          rowDimensionNames.slice(0, 1),
          where,
          sortPivotBy,
          timeRange,
        );

        /**
         * We need to query the unsorted row dimension values because the sorted
         * row dimension values may not have all the dimensions values
         */
        const rowDimensionUnsortedAxisQuery = getAxisForDimensions(
          ctx,
          config,
          rowDimensionNames.slice(0, 1),
          config.whereFilter,
        );

        /**
         * Derive a store from axes queries
         */
        return derived(
          [rowDimensionAxisQuery, rowDimensionUnsortedAxisQuery],
          ([rowDimensionAxes, rowDimensionUnsortedAxis], axesSet) => {
            if (
              rowDimensionAxes?.isFetching ||
              rowDimensionUnsortedAxis?.isFetching
            ) {
              return axesSet({
                isFetching: true,
                data: lastPivotData,
                columnDef: lastPivotColumnDef,
                assembled: false,
                totalColumns,
              });
            }

            const { rows: rowDimensionValues, totals: rowTotals } =
              reconcileMissingDimensionValues(
                anchorDimension,
                rowDimensionAxes,
                rowDimensionUnsortedAxis,
              );

            let columnDef = getColumnDefForPivot(
              config,
              columnDimensionAxes?.data,
            );

            let initialTableCellQuery:
              | Readable<null>
              | CreateQueryResult<V1MetricsViewAggregationResponse, unknown> =
              readable(null);

            if (colDimensionNames.length || !rowDimensionNames.length) {
              const slicedAxesDataForPage = sliceColumnAxesDataForDef(
                colDimensionNames,
                columnDimensionAxes?.data,
                config.pivot.columnPage,
                measureNames.length,
              );

              columnDef = getColumnDefForPivot(config, slicedAxesDataForPage);

              initialTableCellQuery = createTableCellQuery(
                ctx,
                config,
                rowDimensionNames[0],
                columnDimensionAxes?.data,
                rowDimensionValues,
              );
            }

            /**
             * Derive a store from initial table cell data query
             */
            return derived(
              [initialTableCellQuery],
              ([initialTableCellData], cellSet) => {
                let pivotData: PivotDataRow[] = [];
                let cellData: V1MetricsViewAggregationResponseDataItem[] = [];
                if (getPivotConfigKey(config) in expandedTableMap) {
                  pivotData = expandedTableMap[getPivotConfigKey(config)];
                } else {
                  if (initialTableCellData === null) {
                    cellData = rowTotals;
                  } else {
                    if (initialTableCellData.isFetching) {
                      return cellSet({
                        isFetching: true,
                        data: rowTotals,
                        columnDef,
                        assembled: false,
                        totalColumns,
                      });
                    }
                    cellData = initialTableCellData.data?.data || [];
                  }
                  const tableDataWithCells = reduceTableCellDataIntoRows(
                    config,
                    anchorDimension,
                    rowDimensionValues || [],
                    columnDimensionAxes?.data || {},
                    rowTotals,
                    cellData,
                  );

                  pivotData = tableDataWithCells;
                }

                const expandedSubTableCellQuery = queryExpandedRowMeasureValues(
                  ctx,
                  config,
                  pivotData,
                  columnDimensionAxes?.data,
                );
                let globalTotalsQuery:
                  | Readable<null>
                  | CreateQueryResult<
                      V1MetricsViewAggregationResponse,
                      unknown
                    > = readable(null);
                let totalsRowQuery:
                  | Readable<null>
                  | CreateQueryResult<
                      V1MetricsViewAggregationResponse,
                      unknown
                    > = readable(null);
                if (rowDimensionNames.length && measureNames.length) {
                  /** In some cases the totals query would be the same query as that
                   * for the initial table cell data. With svelte query cache we would not hit the
                   * API twice
                   */
                  globalTotalsQuery = createPivotAggregationRowQuery(
                    ctx,
                    config.measureNames,
                    [],
                    config.whereFilter,
                    [],
                    "10000", // Using 10000 for cache hit
                  );
                  totalsRowQuery = createTableCellQuery(
                    ctx,
                    config,
                    undefined,
                    columnDimensionAxes?.data,
                    [],
                  );
                }

                /**
                 * Derive a store based on expanded rows and totals
                 */
                return derived(
                  [
                    globalTotalsQuery,
                    totalsRowQuery,
                    expandedSubTableCellQuery,
                  ],
                  ([
                    globalTotalsResponse,
                    totalsRowResponse,
                    expandedRowMeasureValues,
                  ]) => {
                    prepareNestedPivotData(pivotData, rowDimensionNames);
                    let tableDataExpanded: PivotDataRow[] = pivotData;
                    if (expandedRowMeasureValues?.length) {
                      tableDataExpanded = addExpandedDataToPivot(
                        config,
                        pivotData,
                        rowDimensionNames,
                        columnDimensionAxes?.data || {},
                        expandedRowMeasureValues,
                      );

                      const key = getPivotConfigKey(config);
                      expandedTableMap = {};
                      expandedTableMap[key] = tableDataExpanded;
                    }
                    lastPivotData = tableDataExpanded;
                    lastPivotColumnDef = columnDef;

                    let assembledTableData = tableDataExpanded;
                    if (rowDimensionNames.length && measureNames.length) {
                      const totalsRowData = totalsRowResponse?.data?.data;

                      const globalTotalsData =
                        globalTotalsResponse?.data?.data || [];
                      const totalsRowTable = reduceTableCellDataIntoRows(
                        config,
                        "",
                        [],
                        columnDimensionAxes?.data || {},
                        [],
                        totalsRowData || [],
                      );

                      let totalsRow = totalsRowTable[0] || {};
                      totalsRow[anchorDimension] = "Total";

                      globalTotalsData.forEach((total) => {
                        totalsRow = { ...total, ...totalsRow };
                      });

                      assembledTableData = [totalsRow, ...tableDataExpanded];
                    }

                    return {
                      isFetching: false,
                      data: assembledTableData,
                      columnDef,
                      assembled: true,
                      totalColumns,
                    };
                  },
                ).subscribe(cellSet);
              },
            ).subscribe(axesSet);
          },
        ).subscribe(columnSet);
      },
    ).subscribe(configSet);
  });
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const usePivotDataStore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => createPivotDataStore(ctx),
);
