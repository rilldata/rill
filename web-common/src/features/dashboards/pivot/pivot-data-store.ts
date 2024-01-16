import type {
  V1MetricsViewAggregationSort,
  V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, Readable } from "svelte/store";

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
  getSortForAccessor,
} from "./pivot-utils";
import {
  createTableWithAxes,
  reduceTableCellDataIntoRows,
  prepareNestedPivotData,
} from "./pivot-table-transformations";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
} from "./pivot-queries";
import type { ColumnDef } from "@tanstack/svelte-table";

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

/**
 * Stores the last pivot data and column def to be used when there is no data
 * to be displayed. This is to avoid the table from flickering when there is no
 * data to be displayed.
 */
let lastPivotData: PivotDataRow[] = [];
let lastPivotColumnDef: ColumnDef<PivotDataRow>[] = [];

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
      });
    }
    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
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
          });
        }

        const { filters, sortPivotBy } = getSortForAccessor(
          config,
          columnDimensionAxes?.data,
        );

        const rowDimensionAxisQuery = getAxisForDimensions(
          ctx,
          rowDimensionNames,
          filters,
          sortPivotBy,
        );

        /**
         * Derive a store from axes queries
         */
        return derived(rowDimensionAxisQuery, (rowDimensionAxes, axesSet) => {
          if (rowDimensionAxes?.isFetching) {
            return axesSet({
              isFetching: true,
              data: lastPivotData,
              columnDef: lastPivotColumnDef,
            });
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
                // FIXME: Table does not render properly if below object
                // is set using derived stores set method

                // return cellSet({
                //   isFetching: false,
                //   data: skeletonTableData,
                //   columnDef,
                // });
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
                  lastPivotData = tableDataExpanded;
                  lastPivotColumnDef = columnDef;
                  return {
                    isFetching: false,
                    data: tableDataExpanded,
                    columnDef,
                  };
                },
              ).subscribe(cellSet);
            },
          ).subscribe(axesSet);
        }).subscribe(columnSet);
      },
    ).subscribe(configSet);
  });
}

interface PivotDataState {
  isFetching: boolean;
  data: PivotDataRow[];
  columnDef: ColumnDef<PivotDataRow>[];
}

export type PivotDataStore = Readable<PivotDataState>;

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const usePivotDataStore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => createPivotDataStore(ctx),
);
