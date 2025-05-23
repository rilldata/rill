import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import type { ColumnDef } from "@tanstack/svelte-table";
import { type Readable, derived, readable } from "svelte/store";
import { getColumnDefForPivot } from "./pivot-column-definition";
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
  getTotalsRow,
  getTotalsRowSkeleton,
  mergeRowTotalsInOrder,
  prepareNestedPivotData,
  reduceTableCellDataIntoRows,
} from "./pivot-table-transformations";
import {
  getErrorFromResponses,
  getErrorState,
  getFilterForPivotTable,
  getFiltersForCell,
  getPivotConfigKey,
  getSortFilteredMeasureBody,
  getSortForAccessor,
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
  type PivotFilter,
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
  /**
   * Derive a store using pivot config
   */

  return derived(configStore, (config, configSet) => {
    const { rowDimensionNames, colDimensionNames, measureNames, isFlat } =
      config;

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
        const anchorDimension = rowDimensionNames[0];

        const rowPage = config.pivot.rowPage;
        const rowOffset = (rowPage - 1) * NUM_ROWS_PER_PAGE;

        let whereFilter: V1Expression = config.whereFilter;
        if (config.searchText) {
          whereFilter =
            getDimensionFilterWithSearch(
              whereFilter,
              config.searchText,
              anchorDimension,
            ) || config.whereFilter;
        }

        const {
          where: measureWhere,
          sortPivotBy,
          timeRange,
        } = getSortForAccessor(
          anchorDimension,
          config,
          columnDimensionAxes?.data,
        );

        const { sortFilteredMeasureBody, isMeasureSortAccessor, sortAccessor } =
          getSortFilteredMeasureBody(measureBody, sortPivotBy, measureWhere);

        let rowDimensionAxisQuery: Readable<PivotAxesData | null> =
          readable(null);

        if (!isFlat) {
          // Get sort order for the anchor dimension
          rowDimensionAxisQuery = getAxisForDimensions(
            ctx,
            config,
            rowDimensionNames.slice(0, 1),
            sortFilteredMeasureBody,
            whereFilter,
            sortPivotBy,
            timeRange,
            NUM_ROWS_PER_PAGE.toString(),
            rowOffset.toString(),
          );
        }

        let globalTotalsQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
          readable(null);
        let totalsRowQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
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

        const displayTotalsRow = Boolean(
          rowDimensionNames.length && measureNames.length,
        );
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
                data: lastPivotData,
                columnDef: lastPivotColumnDef,
                assembled: false,
                totalColumns: lastTotalColumns,
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

            const rowDimensionValues =
              rowDimensionAxes?.data?.[anchorDimension] || [];

            const totalColumns = getTotalColumnCount(totalsRowData);

            const axesRowTotals =
              rowDimensionAxes?.totals?.[anchorDimension] || [];

            const rowAxesQueryForMeasureTotals = getAxisQueryForMeasureTotals(
              ctx,
              config,
              isMeasureSortAccessor,
              sortAccessor,
              anchorDimension,
              rowDimensionValues,
              timeRange,
            );

            let tableCellQuery:
              | Readable<null>
              | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
              readable(null);

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
                    data: lastPivotData ? lastPivotData : axesRowTotals,
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

                let pivotSkeleton = mergedRowTotals;
                if (!isFlat && rowPage > 1) {
                  pivotSkeleton = [...lastPivotData, ...mergedRowTotals];
                }

                let pivotData: PivotDataRow[] = [];
                let cellData: PivotDataRow[] = [];
                let isCellDataEmpty = false;
                if (getPivotConfigKey(config) in expandedTableMap) {
                  pivotData = expandedTableMap[getPivotConfigKey(config)];
                } else {
                  if (tableCellData === null) {
                    cellData = pivotSkeleton as PivotDataRow[];
                  } else {
                    if (tableCellData.isFetching) {
                      return cellSet({
                        isFetching: true,
                        data: isFlat ? lastPivotData : pivotSkeleton,
                        columnDef,
                        assembled: false,
                        totalColumns,
                        totalsRowData: displayTotalsRow
                          ? totalsRowData
                          : undefined,
                      });
                    }
                    cellData = (tableCellData.data?.data ||
                      []) as PivotDataRow[];
                    isCellDataEmpty = cellData.length === 0;
                  }

                  let tableDataWithCells: PivotDataRow[] = [];
                  if (isFlat) {
                    if (rowPage > 1) {
                      tableDataWithCells = [...lastPivotData, ...cellData];
                    } else {
                      tableDataWithCells = cellData;
                    }
                  } else {
                    tableDataWithCells = reduceTableCellDataIntoRows(
                      config,
                      anchorDimension,
                      rowDimensionValues || [],
                      columnDimensionAxes?.data || {},
                      pivotSkeleton as PivotDataRow[],
                      cellData,
                    );
                  }
                  pivotData = structuredClone(tableDataWithCells);
                }

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

                      const key = getPivotConfigKey(config);
                      expandedTableMap = {};
                      expandedTableMap[key] = tableDataExpanded;
                    }

                    const activeCell = config.pivot.activeCell;
                    let activeCellFilters: PivotFilter | undefined = undefined;
                    if (activeCell) {
                      activeCellFilters = getFiltersForCell(
                        config,
                        activeCell.rowId,
                        activeCell.columnId,
                        columnDimensionAxes?.data,
                        tableDataExpanded,
                      );
                    }

                    lastPivotData = tableDataExpanded;
                    lastPivotColumnDef = columnDef;
                    lastTotalColumns = totalColumns;

                    const reachedEndForRowData =
                      (isFlat
                        ? isCellDataEmpty
                        : rowDimensionValues.length === 0) && rowPage > 1;
                    return {
                      isFetching: false,
                      data: tableDataExpanded,
                      columnDef,
                      assembled: true,
                      activeCellFilters,
                      totalColumns,
                      reachedEndForRowData,
                      totalsRowData: displayTotalsRow
                        ? totalsRowData
                        : undefined,
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
      metricsViewName: ctx.metricsViewName,
      queryClient: ctx.queryClient,
      enabled: !!ctx.dashboardStore,
    };
    return createPivotDataStore(pivotDashboardContext, pivotConfig);
  },
);
