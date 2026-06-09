import type { ConnectError } from "@connectrpc/connect";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { type Readable, derived, readable } from "svelte/store";
import type { ColumnDef } from "tanstack-table-8-svelte-5";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  assembleBasePivotData,
  buildFinalPivotStateDetails,
  cacheExpandedPivotData,
  createPivotDataCache,
  getPivotSkeletonForPage,
  syncPivotCacheToConfig,
  updatePivotDataCache,
} from "./pivot-data-assembly";
import { getPivotConfig } from "./pivot-data-config";
import {
  addExpandedDataToPivot,
  getExpandedQueryErrors,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import {
  NUM_ROWS_PER_PAGE,
  sliceColumnAxesDataForDef,
} from "./pivot-infinite-scroll";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
  getAxisQueryForMeasureTotals,
  getTotalsRowQuery,
} from "./pivot-queries";
import {
  applyOutermostRowLimit,
  createPivotBaseQueryPlan,
} from "./pivot-query-plan";
import {
  getTotalsRow,
  getTotalsRowSkeleton,
  mergeRowTotalsInOrder,
  prepareNestedPivotData,
} from "./pivot-table-transformations";
import {
  getErrorFromResponses,
  getErrorState,
  getFilterForPivotTable,
  getTimeForQuery,
  getTimeGrainFromDimension,
  getTotalColumnCount,
  isTimeDimension,
  splitPivotChips,
} from "./pivot-utils";
import {
  type PivotAxesData,
  type PivotDashboardContext,
  type PivotDataRow,
  type PivotDataStore,
  type PivotDataStoreConfig,
} from "./types";

/**
 * Returns a query for cell data for the initial table.
 * The table cell is sorted by the anchor dimension irrespective
 * of the sort config. The dimension axes values are sorted using
 * the config and values from this query are used to create the
 * table.
 */
export function createTableCellQuery(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  columnDimensionAxesData: Record<string, string[]> | undefined,
  totalsRow: PivotDataRow,
  rowDimensionValues: string[],
  limit = "5000",
  offset = "0",
) {
  const {
    rowDimensionNames,
    colDimensionNames,
    measureNames,
    isFlat,
    time,
    whereFilter,
  } = config;
  const anchorDimension: string | undefined = rowDimensionNames?.[0];

  const rowPage = config.pivot.rowPage;
  if (!isFlat && rowDimensionValues.length === 0 && rowPage > 1)
    return readable(null);

  let allDimensions = colDimensionNames;

  if (isFlat) {
    allDimensions = colDimensionNames.concat(rowDimensionNames);
  } else if (anchorDimension) {
    allDimensions = colDimensionNames.concat([anchorDimension]);
  }

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
  const measureBody = measureNames.map((m) => ({ name: m }));

  const { filters: filterForInitialTable, timeFilters } =
    getFilterForPivotTable(
      config,
      columnDimensionAxesData,
      totalsRow,
      rowDimensionValues,
      anchorDimension,
    );

  const timeRange: TimeRangeString = getTimeForQuery(time, timeFilters);

  const mergedFilter =
    mergeFilters(filterForInitialTable, whereFilter) ?? createAndExpression([]);

  let sortBy: V1MetricsViewAggregationSort[] = [];
  if (isFlat) {
    const sortConfig = config.pivot.sorting?.[0];
    if (sortConfig) {
      sortBy = [
        {
          desc: sortConfig.desc,
          name: sortConfig.id, // For flat tables, sort ID is directly the measure or dimension name
        },
      ];
    } else {
      // Default sort if no sort config provided
      sortBy = [
        {
          desc: measureNames[0] ? true : false,
          name: measureNames[0] || allDimensions[0],
        },
      ];
    }
  } else {
    sortBy = [
      {
        desc: false,
        name: anchorDimension || measureNames[0],
      },
    ];
  }

  return createPivotAggregationRowQuery(
    ctx,
    config,
    measureBody,
    dimensionBody,
    mergedFilter,
    sortBy,
    limit,
    offset,
    timeRange,
  );
}

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
 * Create skeleton table data by querying axes values for row dimension.
 * Also fetch column wise totals to determine columns to render
 *     |
 *     |  (Cell Data)
 *     v
 * For the visible axes values, query the data for each cell
 *     |
 *     |  (Expanded)
 *     v
 * For each expanded row, query the data for each cell
 *     |
 *     |  (Assemble)
 *     v
 * Table data and column definitions
 */
export function createPivotDataStore(
  ctx: PivotDashboardContext,
  configStore: Readable<PivotDataStoreConfig>,
): PivotDataStore {
  const cache = createPivotDataCache();

  /**
   * Derive a store using pivot config
   */

  return derived(configStore, (config, configSet) => {
    const { rowDimensionNames, colDimensionNames, measureNames, isFlat } =
      config;

    if (config.ready === false) {
      return configSet({
        isFetching: true,
        data: [],
        columnDef: [],
        assembled: false,
        totalColumns: 0,
      });
    }

    if (
      (!rowDimensionNames.length && !measureNames.length) ||
      (colDimensionNames.length && !measureNames.length)
    ) {
      const { dimension: colDimensions, measure: colMeasures } =
        splitPivotChips(config.pivot.columns);
      const isFetching =
        colMeasures.length > 0 ||
        (config.pivot.rows.length > 0 && !colDimensions.length);
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
      [],
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
        if (columnDimensionAxes?.error && columnDimensionAxes?.error.length) {
          return columnSet(getErrorState(columnDimensionAxes.error));
        }
        const plan = createPivotBaseQueryPlan(
          config,
          columnDimensionAxes?.data,
        );
        const { anchorDimension, rowOffset, rowPage } = plan;

        let rowDimensionAxisQuery: Readable<PivotAxesData | null> =
          readable(null);

        if (!isFlat) {
          rowDimensionAxisQuery = getAxisForDimensions(
            ctx,
            config,
            rowDimensionNames.slice(0, 1),
            plan.sortFilteredMeasureBody,
            plan.whereFilter,
            plan.sortPivotBy,
            plan.timeRange,
            plan.rowAxisLimitToQuery,
            rowOffset.toString(),
          );
        }

        let globalTotalsQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, ConnectError> =
          readable(null);
        let totalsRowQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, ConnectError> =
          readable(null);
        if (rowDimensionNames.length && measureNames.length) {
          globalTotalsQuery = createPivotAggregationRowQuery(
            ctx,
            config,
            config.measureNames.map((m) => ({ name: m })),
            [],
            config.whereFilter,
            [],
            "5000", // Using 5000 for cache hit
          );
        }

        const displayTotalsRow = plan.displayTotalsRow;
        if (
          (rowDimensionNames.length || colDimensionNames.length) &&
          measureNames.length &&
          !isFlat
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
              const skeletonTotalsRowData = getTotalsRowSkeleton(
                config,
                columnDimensionAxes?.data,
              );
              return axesSet({
                isFetching: true,
                data: cache.lastPivotData,
                columnDef: cache.lastPivotColumnDef,
                assembled: false,
                totalColumns: cache.lastTotalColumns,
                totalsRowData: displayTotalsRow
                  ? skeletonTotalsRowData
                  : undefined,
              });
            }

            // check for errors in the responses
            const totalErrors = getErrorFromResponses([
              globalTotalsResponse,
              totalsRowResponse,
            ]);

            if (totalErrors.length || rowDimensionAxes?.error?.length) {
              const allErrors = totalErrors.concat(
                rowDimensionAxes?.error || [],
              );
              return axesSet(getErrorState(allErrors));
            }

            /**
             * If there are no axes values, return an empty table
             */
            if (
              (rowDimensionAxes?.data?.[anchorDimension]?.length === 0 ||
                totalsRowResponse?.data?.data?.length === 0) &&
              rowPage === 1
            ) {
              return axesSet({
                isFetching: false,
                data: [],
                columnDef: [],
                assembled: true,
                totalColumns: 0,
                totalsRowData: displayTotalsRow ? [] : undefined,
              });
            }

            const totalsRowData = getTotalsRow(
              config,
              columnDimensionAxes?.data,
              totalsRowResponse?.data?.data,
              globalTotalsResponse?.data?.data,
            );

            const limitedRowAxes = applyOutermostRowLimit(
              config,
              plan,
              rowDimensionAxes?.data?.[anchorDimension] || [],
              rowDimensionAxes?.totals?.[anchorDimension] || [],
            );
            const { axesRowTotals, hasMoreRows, rowDimensionValues } =
              limitedRowAxes;

            const totalColumns = getTotalColumnCount(totalsRowData);

            const rowAxesQueryForMeasureTotals = getAxisQueryForMeasureTotals(
              ctx,
              config,
              plan.isMeasureSortAccessor,
              plan.sortAccessor,
              anchorDimension,
              rowDimensionValues,
              plan.timeRange,
            );

            let tableCellQuery:
              | Readable<null>
              | CreateQueryResult<
                  V1MetricsViewAggregationResponse,
                  ConnectError
                > = readable(null);

            let columnDef: ColumnDef<PivotDataRow>[] = [];
            if (
              isFlat ||
              colDimensionNames.length ||
              !rowDimensionNames.length
            ) {
              const slicedAxesDataForDef = sliceColumnAxesDataForDef(
                config,
                columnDimensionAxes?.data,
                totalsRowData,
              );

              columnDef = getColumnDefForPivot(
                config,
                slicedAxesDataForDef,
                totalsRowData,
              );

              tableCellQuery = createTableCellQuery(
                ctx,
                config,
                columnDimensionAxes?.data,
                totalsRowData,
                rowDimensionValues,
                isFlat ? NUM_ROWS_PER_PAGE.toString() : "5000",
                isFlat ? rowOffset.toString() : "0",
              );
            } else {
              columnDef = getColumnDefForPivot(
                config,
                columnDimensionAxes?.data,
                totalsRowData,
              );
            }
            /**
             * Derive a store from table cell data query
             */
            return derived(
              [rowAxesQueryForMeasureTotals, tableCellQuery],
              ([rowMeasureTotalsAxesQuery, tableCellData], cellSet) => {
                if (rowMeasureTotalsAxesQuery?.isFetching) {
                  return cellSet({
                    isFetching: true,
                    data: cache.lastPivotData.length
                      ? cache.lastPivotData
                      : (axesRowTotals as PivotDataRow[]),
                    columnDef,
                    assembled: false,
                    totalColumns,
                    totalsRowData: displayTotalsRow ? totalsRowData : undefined,
                  });
                }

                const tableCellQueryError = getErrorFromResponses([
                  tableCellData,
                ]);

                if (
                  tableCellQueryError.length ||
                  rowMeasureTotalsAxesQuery?.error?.length
                ) {
                  const allErrors = tableCellQueryError.concat(
                    rowMeasureTotalsAxesQuery?.error || [],
                  );
                  return cellSet(getErrorState(allErrors));
                }

                const mergedRowTotals = mergeRowTotalsInOrder(
                  rowDimensionValues,
                  axesRowTotals,
                  rowMeasureTotalsAxesQuery?.data?.[anchorDimension] || [],
                  rowMeasureTotalsAxesQuery?.totals?.[anchorDimension] || [],
                );

                syncPivotCacheToConfig(cache, plan.configKey);

                const pivotSkeleton = getPivotSkeletonForPage(
                  config,
                  cache,
                  mergedRowTotals as PivotDataRow[],
                );

                let isCellDataEmpty = false;
                let cellData = pivotSkeleton;
                if (tableCellData !== null) {
                  if (tableCellData.isFetching) {
                    return cellSet({
                      isFetching: true,
                      data: isFlat ? cache.lastPivotData : pivotSkeleton,
                      columnDef,
                      assembled: false,
                      totalColumns,
                      totalsRowData: displayTotalsRow
                        ? totalsRowData
                        : undefined,
                    });
                  }
                  cellData = (tableCellData.data?.data || []) as PivotDataRow[];
                  isCellDataEmpty = cellData.length === 0;
                }

                const { pivotData } = assembleBasePivotData({
                  anchorDimension,
                  cache,
                  cellData,
                  columnDimensionAxes: columnDimensionAxes?.data || {},
                  config,
                  configKey: plan.configKey,
                  pivotSkeleton,
                  rowDimensionValues: rowDimensionValues || [],
                });

                const expandedSubTableCellQuery = queryExpandedRowMeasureValues(
                  ctx,
                  config,
                  pivotData,
                  columnDimensionAxes?.data,
                  totalsRowData,
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
                      const queryErrors = getExpandedQueryErrors(
                        expandedRowMeasureValues,
                      );
                      if (queryErrors.length) return getErrorState(queryErrors);

                      tableDataExpanded = addExpandedDataToPivot(
                        config,
                        pivotData,
                        rowDimensionNames,
                        columnDimensionAxes?.data || {},
                        expandedRowMeasureValues,
                      );

                      cacheExpandedPivotData(cache, config, tableDataExpanded);
                    }

                    const finalState = buildFinalPivotStateDetails({
                      anchorDimension,
                      columnDimensionAxes: columnDimensionAxes?.data,
                      config,
                      data: tableDataExpanded,
                      hasMoreRows,
                      isCellDataEmpty,
                      rowDimensionValues,
                      rowOffset,
                    });

                    updatePivotDataCache(cache, {
                      columnDef,
                      data: finalState.data,
                      rowPage,
                      totalColumns,
                    });

                    return {
                      isFetching: false,
                      data: finalState.data,
                      columnDef,
                      assembled: true,
                      activeCellFilters: finalState.activeCellFilters,
                      totalColumns,
                      reachedEndForRowData: finalState.reachedEndForRowData,
                      totalsRowData: displayTotalsRow
                        ? totalsRowData
                        : undefined,
                      columnDimensionAxes: columnDimensionAxes?.data,
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
export const usePivotForExplore = memoizeMetricsStore<PivotDataStore>(
  (ctx: StateManagers) => {
    const pivotConfig = getPivotConfig(ctx);
    const pivotDashboardContext: PivotDashboardContext = {
      runtimeClient: ctx.runtimeClient,
      metricsViewName: ctx.metricsViewName,
      queryClient: ctx.queryClient,
      enabled: readable(!!ctx.dashboardStore),
    };
    return createPivotDataStore(pivotDashboardContext, pivotConfig);
  },
);
