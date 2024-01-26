import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import { derived, Readable } from "svelte/store";

import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { ColumnDef } from "@tanstack/svelte-table";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  addExpandedDataToPivot,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
} from "./pivot-queries";
import {
  prepareNestedPivotData,
  reduceTableCellDataIntoRows,
} from "./pivot-table-transformations";
import {
  getDimensionsInPivotColumns,
  getDimensionsInPivotRow,
  getFilterForPivotTable,
  getMeasuresInPivotColumns,
  getPivotConfigKey,
  getSortForAccessor,
  reconcileMissingDimensionValues,
} from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

/**
 * Extract out config relevant to pivot from dashboard and meta store
 */
function getPivotConfig(ctx: StateManagers): Readable<PivotDataStoreConfig> {
  return derived(
    [useMetaQuery(ctx), ctx.dashboardStore, useTimeControlStore(ctx)],
    ([metricsView, dashboardStore, timeControls]) => {
      const { rows, columns } = dashboardStore.pivot;

      const time = {
        timeStart: timeControls.timeStart,
        timeEnd: timeControls.timeEnd,
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView?.data?.timeDimension || "",
        interval:
          timeControls?.selectedTimeRange?.interval || "TIME_GRAIN_HOUR",
      };

      if (
        (rows.length == 0 && columns.length == 0) ||
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
          filters: dashboardStore.filters,
          pivot: dashboardStore.pivot,
          time,
        };
      }
      const measureNames = getMeasuresInPivotColumns(
        dashboardStore.pivot,
        metricsView.data?.measures,
      );
      const rowDimensionNames = getDimensionsInPivotRow(
        dashboardStore.pivot,
        metricsView.data?.measures,
      );

      const colDimensionNames = getDimensionsInPivotColumns(
        dashboardStore.pivot,
        metricsView.data?.measures,
      );

      return {
        measureNames,
        rowDimensionNames,
        colDimensionNames,
        allMeasures: metricsView.data?.measures,
        allDimensions: metricsView.data?.dimensions,
        filters: dashboardStore.filters,
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
  anchorDimension: string,
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
    dimensionBody,
    filterForInitialTable,
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

    if (!rowDimensionNames.length && !measureNames.length) {
      return configSet({
        isFetching: false,
        data: lastPivotData,
        columnDef: lastPivotColumnDef,
        assembled: false,
      });
    }
    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
      config,
      colDimensionNames,
      config.filters,
    );

    return derived(
      columnDimensionAxesQuery,
      (columnDimensionAxes, columnSet) => {
        if (columnDimensionAxes?.isFetching) {
          return columnSet({
            isFetching: true,
            data: lastPivotData,
            columnDef: lastPivotColumnDef,
            assembled: false,
          });
        }
        const anchorDimension = rowDimensionNames[0];

        const { filters, sortPivotBy, timeRange } = getSortForAccessor(
          anchorDimension,
          config,
          columnDimensionAxes?.data,
        );

        const rowDimensionAxisQuery = getAxisForDimensions(
          ctx,
          config,
          rowDimensionNames.slice(0, 1),
          filters,
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
          config.filters,
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
              });
            }

            const { rows: rowDimensionValues, totals: rowTotals } =
              reconcileMissingDimensionValues(
                anchorDimension,
                rowDimensionAxes,
                rowDimensionUnsortedAxis,
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
              rowDimensionValues,
            );

            /**
             * Derive a store from initial table cell data query
             */
            return derived(
              [initialTableCellQuery],
              ([initialTableCellData], cellSet) => {
                if (initialTableCellData.isFetching) {
                  return cellSet({
                    isFetching: true,
                    data: rowTotals,
                    columnDef,
                    assembled: false,
                  });
                }

                const cellData = initialTableCellData.data
                  ?.data as V1MetricsViewAggregationResponseDataItem[];

                const tableDataWithCells = reduceTableCellDataIntoRows(
                  config,
                  anchorDimension,
                  rowDimensionValues || [],
                  columnDimensionAxes?.data || {},
                  rowTotals,
                  cellData,
                );

                let pivotData = tableDataWithCells;

                // TODO: Considering optimizing this derived store
                if (getPivotConfigKey(config) in expandedTableMap) {
                  pivotData = expandedTableMap[getPivotConfigKey(config)];
                }

                const expandedSubTableCellQuery = queryExpandedRowMeasureValues(
                  ctx,
                  config,
                  pivotData,
                  columnDimensionAxes?.data,
                );
                /** In some cases the totals query would be the same query as that
                 * for the initial table cell data. With svelte query cache we would not hit the
                 * API twice
                 */
                const totalsRowQuery = createTableCellQuery(
                  ctx,
                  config,
                  "",
                  columnDimensionAxes?.data,
                  [],
                );

                /**
                 * Derive a store based on expanded rows and totals
                 */
                return derived(
                  [totalsRowQuery, expandedSubTableCellQuery],
                  ([totalsRowResponse, expandedRowMeasureValues]) => {
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
                    if (config.rowDimensionNames.length) {
                      const totalsRowData = totalsRowResponse?.data?.data;
                      const totalsRowTable = reduceTableCellDataIntoRows(
                        config,
                        "",
                        [],
                        columnDimensionAxes?.data || {},
                        [],
                        totalsRowData || [],
                      );

                      const totalsRow = totalsRowTable[0] || {};
                      totalsRow[anchorDimension] = "Total";

                      assembledTableData = [totalsRow, ...tableDataExpanded];
                    }

                    return {
                      isFetching: false,
                      data: assembledTableData,
                      columnDef,
                      assembled: true,
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

interface PivotDataState {
  isFetching: boolean;
  data: PivotDataRow[];
  columnDef: ColumnDef<PivotDataRow>[];
  assembled: boolean;
}

export type PivotDataStore = Readable<PivotDataState>;

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const usePivotDataStore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => createPivotDataStore(ctx),
);
