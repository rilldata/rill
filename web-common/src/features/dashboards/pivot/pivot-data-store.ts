import type { ConnectError } from "@connectrpc/connect";
import { getURIRequestMeasure } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { memoizeMetricsStore } from "@rilldata/web-common/features/dashboards/state-managers/memoize-metrics-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { type Readable, derived, readable } from "svelte/store";
import type { ColumnDef } from "tanstack-table-8-svelte-5";
import { getColumnDefForPivot } from "./pivot-column-definition";
import {
  type PivotDataCache,
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
  type LimitedRowAxes,
  type PivotBaseQueryPlan,
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
  type PivotDataState,
  type PivotDataStore,
  type PivotDataStoreConfig,
} from "./types";

type AggregationQuery =
  | Readable<null>
  | CreateQueryResult<V1MetricsViewAggregationResponse, ConnectError>;

/**
 * Context threaded through the query stages of the pivot data pipeline.
 * Each stage's args extend the previous stage's args with the values that
 * stage resolved.
 */
interface RowAxesStageArgs {
  ctx: PivotDashboardContext;
  config: PivotDataStoreConfig;
  cache: PivotDataCache;
  plan: PivotBaseQueryPlan;
  columnDimensionAxes: Record<string, string[]> | undefined;
}

interface CellDataStageArgs extends RowAxesStageArgs {
  totalsRowData: PivotDataRow;
  limitedRowAxes: LimitedRowAxes;
  totalColumns: number;
}

interface ExpandedDataStageArgs extends CellDataStageArgs {
  columnDef: ColumnDef<PivotDataRow>[];
  pivotData: PivotDataRow[];
  isCellDataEmpty: boolean;
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
 *
 * Each step depends on the results of the previous one, so the pipeline is
 * a chain of stage functions linked with `derivedSwitch`: a stage either
 * exits early with a loading/error/empty state or hands off to the next
 * stage's store.
 */
export function createPivotDataStore(
  ctx: PivotDashboardContext,
  configStore: Readable<PivotDataStoreConfig>,
): PivotDataStore {
  const cache = createPivotDataCache();

  return derivedSwitch(configStore, (config) => {
    const { rowDimensionNames, colDimensionNames, measureNames } = config;

    if (config.ready === false) {
      return getEmptyState(true);
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
      return getEmptyState(isFetching);
    }

    return createColumnAxesStage(ctx, config, cache);
  });
}

/**
 * Stage 1: query axes values for each column dimension to create table
 * headers, then build the query plan shared by the remaining stages.
 */
function createColumnAxesStage(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  cache: PivotDataCache,
): Readable<PivotDataState> {
  const measureBody = config.measureNames.map((m) => ({ name: m }));

  const columnAxesQuery = getAxisForDimensions(
    ctx,
    config,
    config.colDimensionNames,
    measureBody,
    config.whereFilter,
    [],
  );

  return derivedSwitch(columnAxesQuery, (columnAxesResult) => {
    if (columnAxesResult?.isFetching) {
      return getEmptyState(true);
    }
    if (columnAxesResult?.error?.length) {
      return getErrorState(columnAxesResult.error);
    }

    return createRowAxesStage({
      ctx,
      config,
      cache,
      plan: createPivotBaseQueryPlan(config, columnAxesResult?.data),
      columnDimensionAxes: columnAxesResult?.data,
    });
  });
}

/**
 * Stage 2: query axes values for the row dimension to create skeleton table
 * data and sort order, along with global and column wise totals.
 */
function createRowAxesStage(args: RowAxesStageArgs): Readable<PivotDataState> {
  const { ctx, config, cache, plan, columnDimensionAxes } = args;
  const { rowDimensionNames, colDimensionNames, measureNames, isFlat } = config;

  let rowAxesQuery: Readable<PivotAxesData | null> = readable(null);
  if (!isFlat) {
    rowAxesQuery = getAxisForDimensions(
      ctx,
      config,
      rowDimensionNames.slice(0, 1),
      plan.sortFilteredMeasureBody,
      plan.whereFilter,
      plan.sortPivotBy,
      plan.timeRange,
      plan.rowAxisLimitToQuery,
      plan.rowOffset.toString(),
    );
  }

  let globalTotalsQuery: AggregationQuery = readable(null);
  if (plan.displayTotalsRow) {
    globalTotalsQuery = createPivotAggregationRowQuery(
      ctx,
      config,
      plan.measureBody,
      [],
      config.whereFilter,
      [],
      "5000", // Using 5000 for cache hit
    );
  }

  let totalsRowQuery: AggregationQuery = readable(null);
  if (
    (rowDimensionNames.length || colDimensionNames.length) &&
    measureNames.length &&
    !isFlat
  ) {
    totalsRowQuery = getTotalsRowQuery(ctx, config, columnDimensionAxes);
  }

  return derivedSwitch(
    [rowAxesQuery, globalTotalsQuery, totalsRowQuery],
    ([rowAxesResult, globalTotalsResponse, totalsRowResponse]) => {
      if (
        (globalTotalsResponse !== null && globalTotalsResponse?.isFetching) ||
        (totalsRowResponse !== null && totalsRowResponse?.isFetching) ||
        rowAxesResult?.isFetching
      ) {
        return {
          isFetching: true,
          data: cache.lastPivotData,
          columnDef: cache.lastPivotColumnDef,
          assembled: false,
          totalColumns: cache.lastTotalColumns,
          totalsRowData: plan.displayTotalsRow
            ? getTotalsRowSkeleton(config, columnDimensionAxes)
            : undefined,
        };
      }

      // check for errors in the responses
      const totalErrors = getErrorFromResponses([
        globalTotalsResponse,
        totalsRowResponse,
      ]);
      if (totalErrors.length || rowAxesResult?.error?.length) {
        return getErrorState(totalErrors.concat(rowAxesResult?.error || []));
      }

      // If there are no axes values, return an empty table
      if (
        (rowAxesResult?.data?.[plan.anchorDimension]?.length === 0 ||
          totalsRowResponse?.data?.data?.length === 0) &&
        plan.rowPage === 1
      ) {
        return {
          isFetching: false,
          data: [],
          columnDef: [],
          assembled: true,
          totalColumns: 0,
          totalsRowData: plan.displayTotalsRow ? {} : undefined,
        };
      }

      const totalsRowData = getTotalsRow(
        config,
        columnDimensionAxes,
        totalsRowResponse?.data?.data,
        globalTotalsResponse?.data?.data,
      );

      const limitedRowAxes = applyOutermostRowLimit(
        config,
        plan,
        rowAxesResult?.data?.[plan.anchorDimension] || [],
        rowAxesResult?.totals?.[plan.anchorDimension] || [],
      );

      return createCellDataStage({
        ...args,
        totalsRowData,
        limitedRowAxes,
        totalColumns: getTotalColumnCount(totalsRowData),
      });
    },
  );
}

/**
 * Stage 3: for the visible axes values, query measure totals per row and the
 * cell data for the table body.
 */
function createCellDataStage(
  args: CellDataStageArgs,
): Readable<PivotDataState> {
  const {
    ctx,
    config,
    cache,
    plan,
    columnDimensionAxes,
    totalsRowData,
    limitedRowAxes,
    totalColumns,
  } = args;
  const { axesRowTotals, rowDimensionValues } = limitedRowAxes;
  const { rowDimensionNames, colDimensionNames, isFlat } = config;

  const rowMeasureTotalsQuery = getAxisQueryForMeasureTotals(
    ctx,
    config,
    plan.isMeasureSortAccessor,
    plan.sortAccessor,
    plan.anchorDimension,
    rowDimensionValues,
    plan.timeRange,
  );

  let tableCellQuery: AggregationQuery = readable(null);
  let columnDef: ColumnDef<PivotDataRow>[] = [];
  if (isFlat || colDimensionNames.length || !rowDimensionNames.length) {
    const slicedAxesDataForDef = sliceColumnAxesDataForDef(
      config,
      columnDimensionAxes,
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
      columnDimensionAxes,
      totalsRowData,
      rowDimensionValues,
      isFlat ? NUM_ROWS_PER_PAGE.toString() : "5000",
      isFlat ? plan.rowOffset.toString() : "0",
    );
  } else {
    columnDef = getColumnDefForPivot(
      config,
      columnDimensionAxes,
      totalsRowData,
    );
  }

  return derivedSwitch(
    [rowMeasureTotalsQuery, tableCellQuery],
    ([rowMeasureTotalsResult, tableCellResponse]) => {
      if (rowMeasureTotalsResult?.isFetching) {
        return {
          isFetching: true,
          data: cache.lastPivotData.length
            ? cache.lastPivotData
            : (axesRowTotals as PivotDataRow[]),
          columnDef,
          assembled: false,
          totalColumns,
          totalsRowData: plan.displayTotalsRow ? totalsRowData : undefined,
        };
      }

      const tableCellQueryError = getErrorFromResponses([tableCellResponse]);
      if (tableCellQueryError.length || rowMeasureTotalsResult?.error?.length) {
        return getErrorState(
          tableCellQueryError.concat(rowMeasureTotalsResult?.error || []),
        );
      }

      const mergedRowTotals = mergeRowTotalsInOrder(
        rowDimensionValues,
        axesRowTotals,
        rowMeasureTotalsResult?.data?.[plan.anchorDimension] || [],
        rowMeasureTotalsResult?.totals?.[plan.anchorDimension] || [],
      );

      syncPivotCacheToConfig(cache, plan.configKey);

      const pivotSkeleton = getPivotSkeletonForPage(
        config,
        cache,
        mergedRowTotals as PivotDataRow[],
      );

      let isCellDataEmpty = false;
      let cellData = pivotSkeleton;
      // When cached expanded data exists for this config, keep showing it
      // instead of a loading state while the cell query refetches
      const hasCachedExpandedData = plan.configKey in cache.expandedTableMap;
      if (!hasCachedExpandedData && tableCellResponse !== null) {
        if (tableCellResponse.isFetching) {
          return {
            isFetching: true,
            data: isFlat ? cache.lastPivotData : pivotSkeleton,
            columnDef,
            assembled: false,
            totalColumns,
            totalsRowData: plan.displayTotalsRow ? totalsRowData : undefined,
          };
        }
        cellData = (tableCellResponse.data?.data || []) as PivotDataRow[];
        isCellDataEmpty = cellData.length === 0;
      }

      const { pivotData } = assembleBasePivotData({
        anchorDimension: plan.anchorDimension,
        cache,
        cellData,
        columnDimensionAxes: columnDimensionAxes || {},
        config,
        configKey: plan.configKey,
        pivotSkeleton,
        rowDimensionValues: rowDimensionValues || [],
      });

      return createExpandedDataStage({
        ...args,
        columnDef,
        pivotData,
        isCellDataEmpty,
      });
    },
  );
}

/**
 * Stage 4: query measure values for each expanded row, then assemble the
 * final table data and column definitions.
 */
function createExpandedDataStage(
  args: ExpandedDataStageArgs,
): Readable<PivotDataState> {
  const {
    ctx,
    config,
    cache,
    plan,
    columnDimensionAxes,
    totalsRowData,
    limitedRowAxes,
    totalColumns,
    columnDef,
    pivotData,
    isCellDataEmpty,
  } = args;

  const expandedRowMeasureValuesQuery = queryExpandedRowMeasureValues(
    ctx,
    config,
    pivotData,
    columnDimensionAxes,
    totalsRowData,
  );

  return derived(expandedRowMeasureValuesQuery, (expandedRowMeasureValues) => {
    prepareNestedPivotData(pivotData, config.rowDimensionNames);
    let tableDataExpanded: PivotDataRow[] = pivotData;
    if (expandedRowMeasureValues?.length) {
      const queryErrors = getExpandedQueryErrors(expandedRowMeasureValues);
      if (queryErrors.length) return getErrorState(queryErrors);

      tableDataExpanded = addExpandedDataToPivot(
        config,
        pivotData,
        config.rowDimensionNames,
        columnDimensionAxes || {},
        expandedRowMeasureValues,
      );

      cacheExpandedPivotData(cache, config, tableDataExpanded);
    }

    const { data, activeCellFilters, reachedEndForRowData } =
      buildFinalPivotStateDetails({
        anchorDimension: plan.anchorDimension,
        columnDimensionAxes,
        config,
        data: tableDataExpanded,
        hasMoreRows: limitedRowAxes.hasMoreRows,
        isCellDataEmpty,
        rowDimensionValues: limitedRowAxes.rowDimensionValues,
        rowOffset: plan.rowOffset,
      });

    updatePivotDataCache(cache, {
      columnDef,
      data,
      rowPage: plan.rowPage,
      totalColumns,
    });

    return {
      isFetching: false,
      data,
      columnDef,
      assembled: true,
      activeCellFilters,
      totalColumns,
      reachedEndForRowData,
      totalsRowData: plan.displayTotalsRow ? totalsRowData : undefined,
      columnDimensionAxes,
    };
  });
}

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
  const measureBody: V1MetricsViewAggregationMeasure[] = measureNames.map(
    (m) => ({ name: m }),
  );

  // Request computed URI measures for flat-mode row dimensions that define a
  // uri template, so dimension cells can render as links (matching leaderboard).
  if (isFlat) {
    for (const dimensionName of rowDimensionNames) {
      const dimSpec = config.allDimensions.find(
        (d) => d.name === dimensionName || d.column === dimensionName,
      );
      if (dimSpec?.uri && dimSpec.name) {
        measureBody.push(getURIRequestMeasure(dimSpec.name));
      }
    }
  }

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

function getEmptyState(isFetching: boolean): PivotDataState {
  return {
    isFetching,
    data: [],
    columnDef: [],
    assembled: false,
    totalColumns: 0,
  };
}

type Stores =
  | Readable<unknown>
  | [Readable<unknown>, ...Array<Readable<unknown>>]
  | Array<Readable<unknown>>;
type StoresValues<T> =
  T extends Readable<infer U>
    ? U
    : { [K in keyof T]: T[K] extends Readable<infer U> ? U : never };

/**
 * Like `derived`, but the callback can return either a plain value or
 * another store. A returned store is subscribed to and its values are
 * forwarded, so stages of dependent queries chain linearly instead of
 * nesting derived stores inside each other.
 */
function derivedSwitch<S extends Stores, T>(
  stores: S,
  fn: (values: StoresValues<S>) => T | Readable<T>,
): Readable<T> {
  return derived(stores, (values: StoresValues<S>, set: (value: T) => void) => {
    const result = fn(values);
    if (isReadable(result)) return result.subscribe(set);
    set(result);
  });
}

function isReadable<T>(value: T | Readable<T>): value is Readable<T> {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as Readable<T>).subscribe === "function"
  );
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
