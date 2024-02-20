import { measureFilterResolutionsStore } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query/build/lib/types";
import type { ColumnDef } from "@tanstack/svelte-table";
import { Readable, derived, readable } from "svelte/store";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  addExpandedDataToPivot,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import {
  NUM_ROWS_PER_PAGE,
  sliceColumnAxesDataForDef,
} from "./pivot-infinite-scroll";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
  getTotalsRowQuery,
} from "./pivot-queries";
import {
  getTotalsRow,
  prepareNestedPivotData,
  reduceTableCellDataIntoRows,
} from "./pivot-table-transformations";
import {
  getFilterForPivotTable,
  getPivotConfigKey,
  getSortForAccessor,
  getTimeForQuery,
  getTimeGrainFromDimension,
  getTotalColumnCount,
  isTimeDimension,
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
    [
      useMetricsView(ctx),
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      measureFilterResolutionsStore(ctx),
    ],
    ([metricsView, dashboardStore, timeControls, measureFilterResolution]) => {
      const time: PivotTimeConfig = {
        timeStart: timeControls.timeStart,
        timeEnd: timeControls.timeEnd,
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView?.data?.timeDimension || "",
      };

      if (
        !metricsView.data?.measures ||
        !metricsView.data?.dimensions ||
        !timeControls.ready ||
        !measureFilterResolution.ready
      ) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          whereFilter: dashboardStore.whereFilter,
          measureFilter: measureFilterResolution,
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
          return `${time.timeDimension}_rill_${d.id}`;
        }
        return d.id;
      });

      const colDimensionNames = dashboardStore.pivot.columns.dimension.map(
        (d) => {
          if (d.type === PivotChipType.Time) {
            return `${time.timeDimension}_rill_${d.id}`;
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
        measureFilter: measureFilterResolution,
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
  totalsRow: PivotDataRow,
  rowDimensionValues: string[],
) {
  let allDimensions = config.colDimensionNames;
  if (anchorDimension) {
    allDimensions = config.colDimensionNames.concat([anchorDimension]);
  }

  const { time } = config;
  const dimensionBody = allDimensions.map((dimension) => {
    if (isTimeDimension(dimension, time.timeDimension)) {
      return {
        name: time.timeDimension,
        timeGrain: getTimeGrainFromDimension(dimension),
        timeZone: time.timeZone,
        alias: dimension,
      };
    } else return { name: dimension };
  });
  const measureBody = config.measureNames.map((m) => ({ name: m }));

  const { filters: filterForInitialTable, timeFilters } =
    getFilterForPivotTable(
      config,
      columnDimensionAxesData,
      totalsRow,
      rowDimensionValues,
      true,
      anchorDimension,
    );

  const timeRange: TimeRangeString = getTimeForQuery(config.time, timeFilters);

  const mergedFilter = mergeFilters(filterForInitialTable, config.whereFilter);

  const sortBy = [
    {
      desc: false,
      name: anchorDimension || config.measureNames[0],
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    measureBody,
    dimensionBody,
    mergedFilter,
    config.measureFilter,
    sortBy,
    "5000",
    "0",
    timeRange,
  );
}

/**
 * Stores the last pivot data and column def to be used when there is no data
 * to be displayed. This is to avoid the table from flickering when there is no
 * data to be displayed.
 */
let lastPivotData: PivotDataRow[] = [];
let lastPivotColumnDef: ColumnDef<PivotDataRow>[] = [];
let lastTotalColumns: number = 0;

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
      const isFetching =
        config.pivot.columns.measure.length > 0 ||
        config.pivot.rows.dimension.length > 0;
      return configSet({
        isFetching: isFetching,
        data: [],
        columnDef: [],
        assembled: false,
        totalColumns: 0,
      });
    }
    const measureBody = measureNames.map((m) => ({ name: m }));

    const columnDimensionAxesQuery = getAxisForDimensions(
      ctx,
      config,
      colDimensionNames,
      measureBody,
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

        const {
          where: measureWhere,
          sortPivotBy,
          timeRange,
        } = getSortForAccessor(
          anchorDimension,
          config,
          columnDimensionAxes?.data,
        );

        let sortFilteredMeasureBody = measureBody;
        if (sortPivotBy.length && measureWhere) {
          const accessor = sortPivotBy[0]?.name;
          sortFilteredMeasureBody = measureBody.map((m) => {
            if (m.name === accessor) return { ...m, filter: measureWhere };
            return m;
          });
        }

        const rowOffset = (config.pivot.rowPage - 1) * NUM_ROWS_PER_PAGE;
        const rowDimensionAxisQuery = getAxisForDimensions(
          ctx,
          config,
          rowDimensionNames.slice(0, 1),
          sortFilteredMeasureBody,
          config.whereFilter,
          sortPivotBy,
          timeRange,
          NUM_ROWS_PER_PAGE.toString(),
          rowOffset.toString(),
        );

        let globalTotalsQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, unknown> =
          readable(null);
        let totalsRowQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, unknown> =
          readable(null);
        if (rowDimensionNames.length && measureNames.length) {
          globalTotalsQuery = createPivotAggregationRowQuery(
            ctx,
            config.measureNames.map((m) => ({ name: m })),
            [],
            config.whereFilter,
            config.measureFilter,
            [],
            "5000", // Using 5000 for cache hit
          );
        }
        if (
          (rowDimensionNames.length || colDimensionNames.length) &&
          measureNames.length
        ) {
          totalsRowQuery = getTotalsRowQuery(
            ctx,
            config,
            columnDimensionAxes?.data,
          );
        }

        /**
         * Derive a store from axes queries
         */
        return derived(
          [rowDimensionAxisQuery, globalTotalsQuery, totalsRowQuery],
          (
            [rowDimensionAxes, globalTotalsResponse, totalsRowResponse],
            axesSet,
          ) => {
            if (
              (globalTotalsResponse !== null &&
                globalTotalsResponse?.isFetching) ||
              (totalsRowResponse !== null && totalsRowResponse?.isFetching) ||
              rowDimensionAxes?.isFetching
            ) {
              return axesSet({
                isFetching: true,
                data: lastPivotData,
                columnDef: lastPivotColumnDef,
                assembled: false,
                totalColumns: lastTotalColumns,
              });
            }

            const rowDimensionValues =
              rowDimensionAxes?.data?.[anchorDimension] || [];
            const rowTotals = rowDimensionAxes?.totals?.[anchorDimension] || [];
            const totalsRow = getTotalsRow(
              config,
              columnDimensionAxes?.data,
              totalsRowResponse?.data?.data,
              globalTotalsResponse?.data?.data,
            );

            const totalColumns = getTotalColumnCount(totalsRow);

            let initialTableCellQuery:
              | Readable<null>
              | CreateQueryResult<V1MetricsViewAggregationResponse, unknown> =
              readable(null);

            let columnDef: ColumnDef<PivotDataRow>[] = [];
            if (colDimensionNames.length || !rowDimensionNames.length) {
              const slicedAxesDataForDef = sliceColumnAxesDataForDef(
                config,
                columnDimensionAxes?.data,
                totalsRow,
              );

              columnDef = getColumnDefForPivot(
                config,
                slicedAxesDataForDef,
                totalsRow,
              );

              initialTableCellQuery = createTableCellQuery(
                ctx,
                config,
                rowDimensionNames[0],
                columnDimensionAxes?.data,
                totalsRow,
                rowDimensionValues,
              );
            } else {
              columnDef = getColumnDefForPivot(
                config,
                columnDimensionAxes?.data,
                totalsRow,
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
                  console.log("pivotData", pivotData);
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
                  pivotData = structuredClone(tableDataWithCells);
                }

                const expandedSubTableCellQuery = queryExpandedRowMeasureValues(
                  ctx,
                  config,
                  pivotData,
                  columnDimensionAxes?.data,
                  totalsRow,
                );

                /**
                 * Derive a store based on expanded rows and totals
                 */
                return derived(
                  [expandedSubTableCellQuery],
                  ([expandedRowMeasureValues]) => {
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
                    lastTotalColumns = totalColumns;

                    let assembledTableData = tableDataExpanded;
                    if (rowDimensionNames.length && measureNames.length) {
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
