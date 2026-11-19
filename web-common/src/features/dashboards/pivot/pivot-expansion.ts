import {
  LOADING_CELL,
  MAX_ROW_EXPANSION_LIMIT,
  PIVOT_TOTALS_ROW_ID,
  SHOW_MORE_BUTTON,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  encodeExpandKeySegment,
  expandKeyDepth,
  expandKeySegments,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-expand-keys";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { ConnectError } from "@connectrpc/connect";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { type Readable, derived, readable, writable } from "svelte/store";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
  getAxisQueryForMeasureTotals,
} from "./pivot-queries";
import {
  mergeRowTotalsInOrder,
  reduceTableCellDataIntoRows,
} from "./pivot-table-transformations";
import {
  getErrorFromResponses,
  getFilterForPivotTable,
  getSortFilteredMeasureBody,
  getSortForAccessor,
  getTimeForQuery,
  getTimeGrainFromDimension,
  isTimeDimension,
  mergeTimeStrings,
} from "./pivot-utils";
import type {
  PivotDashboardContext,
  PivotDataRow,
  PivotDataStoreConfig,
  PivotQueryError,
  TimeFilters,
} from "./types";

/**
 * Returns the rows along the resolvable part of a (dimension, value) path.
 * Callers that need a full match must compare its length with the key depth.
 */
function getRowsForExpandedKey(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  key: string,
): PivotDataRow[] {
  const rows: PivotDataRow[] = [];
  // The totals row is not part of the table data and has no dimension values.
  if (key === PIVOT_TOTALS_ROW_ID) return rows;

  // Each row stores its own value under rowDimensions[0] at every depth, so
  // match key segments against that to resolve a key to the actual values.
  const anchor = rowDimensions[0];
  let currentRows: PivotDataRow[] | undefined = tableData;

  for (const [depth, segment] of expandKeySegments(key).entries()) {
    const node: PivotDataRow | undefined = currentRows?.find(
      (row) =>
        encodeExpandKeySegment(rowDimensions[depth], row[anchor]) === segment,
    );
    if (!node) break;
    rows.push(node);
    currentRows = node.subRows;
  }
  return rows;
}

export function getValuesForExpandedKey(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  key: string,
): string[] {
  return getRowsForExpandedKey(tableData, rowDimensions, key).map(
    (row) => row[rowDimensions[0]] as string,
  );
}

/**
 * Returns the dimension values along an expanded key's path when the whole
 * path is in the table data and there is a dimension below it to expand into.
 * Otherwise returns undefined: the key is either waiting for its parent's sub
 * rows to load, or stale because its dimensions no longer match the rows.
 */
function resolveExpandedKey(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  key: string,
): string[] | undefined {
  const depth = expandKeyDepth(key);
  if (depth >= rowDimensions.length) return undefined;
  const values = getValuesForExpandedKey(tableData, rowDimensions, key);
  if (
    values.length !== depth ||
    values.some((v) => v === undefined || v === LOADING_CELL)
  ) {
    return undefined;
  }
  return values;
}

/**
 * Counts the expanded keys that resolve against the table data. Filling in a
 * parent's sub rows can make its expanded children resolvable, which raises
 * the count.
 */
export function countResolvedExpandedKeys(
  config: PivotDataStoreConfig,
  tableData: PivotDataRow[],
): number {
  return Object.keys(config.pivot.expanded).filter(
    (key) =>
      resolveExpandedKey(tableData, config.rowDimensionNames, key) !==
      undefined,
  ).length;
}

/**
 * Returns a query for cell data for a sub table. The values are
 * sorted by anchor dimension so that then can be reduced into
 * rows optimally.
 */
export function createSubTableCellQuery(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  anchorDimension: string,
  columnDimensionAxesData: Record<string, string[]> | undefined,
  totalsRow: PivotDataRow,
  rowDimensionValues: string[],
  rowNestFilters: V1Expression,
  timeFilters: TimeFilters[],
) {
  // A row with no sub rows has no cells to fetch, and its row filter would be
  // an `IN` with no values, which the runtime rejects.
  if (!rowDimensionValues.length) return readable(null);

  const allDimensions = config.colDimensionNames.concat([anchorDimension]);

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

  const { filters: filterForSubTable, timeFilters: colTimeFilters } =
    getFilterForPivotTable(
      config,
      columnDimensionAxesData,
      totalsRow,
      rowDimensionValues,
      anchorDimension,
    );

  const timeRange: TimeRangeString = getTimeForQuery(
    time,
    timeFilters.concat(colTimeFilters),
  );

  filterForSubTable.cond?.exprs?.push(...(rowNestFilters?.cond?.exprs ?? []));

  const sortBy = [
    {
      desc: false,
      name: anchorDimension,
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    config,
    measureBody,
    dimensionBody,
    filterForSubTable,
    sortBy,
    "5000",
    "0",
    timeRange,
  );
}

interface ExpandedRowMeasureValues {
  isFetching: boolean;
  expandIndex: string;
  error?: PivotQueryError[];
  rowDimensionValues: string[];
  data: V1MetricsViewAggregationResponseDataItem[];
  totals: V1MetricsViewAggregationResponseDataItem[];
  hasMore?: boolean; // True if we got limit+1 results (indicates more data available)
  effectiveLimit?: number; // The limit that was applied to this query
}

export function getExpandedErrorState(
  errors: PivotQueryError[],
  expandIndex: string,
): ExpandedRowMeasureValues {
  return {
    isFetching: false,
    error: errors,
    expandIndex,
    rowDimensionValues: [],
    totals: [],
    data: [],
  };
}

/**
 * For each expanded row, create a query for the sub table
 * and return the query result along with the expanded row index
 * and the row dimension values
 */
export function queryExpandedRowMeasureValues(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  tableData: PivotDataRow[],
  columnDimensionAxesData: Record<string, string[]> | undefined,
  totalsRow: PivotDataRow,
): Readable<ExpandedRowMeasureValues[] | null> {
  const { rowDimensionNames } = config;
  const expanded = config.pivot.expanded;
  if (!tableData || Object.keys(expanded).length == 0) return readable(null);

  const numMeasures = config.measureNames.length;
  let measureBody: V1MetricsViewAggregationMeasure[] = config.measureNames.map(
    (m) => ({ name: m }),
  );

  if (numMeasures === 0) {
    measureBody = [
      { name: "__count", builtinMeasure: "BUILTIN_MEASURE_COUNT" },
    ];
  }
  return derived(
    Object.keys(expanded)?.map((expandIndex) => {
      const nestLevel = expandKeyDepth(expandIndex);

      if (nestLevel >= rowDimensionNames.length)
        return readable({
          isFetching: false,
          expandIndex,
          rowDimensionValues: [],
          data: [],
          totals: [],
        });
      const anchorDimension = rowDimensionNames[nestLevel];
      const values = resolveExpandedKey(
        tableData,
        rowDimensionNames,
        expandIndex,
      );

      // Don't query with a partial filter. The data store reruns this stage
      // once a parent's sub rows fill in and make more keys resolvable.
      if (!values)
        return readable({
          isFetching: true,
          expandIndex,
          rowDimensionValues: [],
          data: [],
          totals: [],
        });

      const rowNestTimeFilters: TimeFilters[] = [];
      const rowNestFilters = values
        .map((value, index) =>
          createInExpression(rowDimensionNames[index], [value]),
        )
        .filter((f) => {
          // We map first and filter later to ensure that dimensions are in order
          if (
            isTimeDimension(f.cond?.exprs?.[0].ident, config.time.timeDimension)
          ) {
            rowNestTimeFilters.push({
              timeStart: f.cond?.exprs?.[1].val as string,
              interval: getTimeGrainFromDimension(
                f.cond?.exprs?.[0].ident as string,
              ),
            });
            return false;
          } else return true;
        });

      let filterForRowDimensionAxes: V1Expression | undefined = undefined;
      if (rowNestFilters.length) {
        filterForRowDimensionAxes = createAndExpression(rowNestFilters);
      }

      const {
        where: measureWhere,
        sortPivotBy,
        timeRange: timeRangeSortedCol,
      } = getSortForAccessor(anchorDimension, config, columnDimensionAxesData);

      const mergeRowAndSortFilters = mergeFilters(
        filterForRowDimensionAxes,
        measureWhere,
      );

      const timeRangeRow: TimeRangeString = getTimeForQuery(
        config.time,
        rowNestTimeFilters,
      );
      const timeRange = mergeTimeStrings(timeRangeRow, timeRangeSortedCol);

      const { sortFilteredMeasureBody, isMeasureSortAccessor, sortAccessor } =
        getSortFilteredMeasureBody(
          measureBody,
          sortPivotBy,
          mergeRowAndSortFilters,
        );

      const subTableMergedFilters =
        mergeFilters(filterForRowDimensionAxes, config.whereFilter) ??
        createAndExpression([]);

      // Get effective limit (local limit overrides global limit)
      const localLimit = config.pivot.nestedRowLimits?.[expandIndex];
      const effectiveLimit =
        localLimit ?? config.pivot.rowLimit ?? MAX_ROW_EXPANSION_LIMIT;

      // Query for limit+1 to detect if more data is available
      const queryLimit = (effectiveLimit + 1).toString();

      return derived(
        [
          writable(expandIndex),
          getAxisForDimensions(
            ctx,
            config,
            [anchorDimension],
            sortFilteredMeasureBody,
            subTableMergedFilters,
            sortPivotBy,
            timeRange,
            queryLimit,
            "0",
            true,
          ),
        ],
        ([expandIndex, subRowDimensions], axisSet) => {
          if (subRowDimensions?.error?.length) {
            return axisSet(
              getExpandedErrorState(subRowDimensions.error, expandIndex),
            );
          }
          if (subRowDimensions?.isFetching) {
            const rowMeasureValuesEmpty: ExpandedRowMeasureValues = {
              isFetching: true,
              expandIndex,
              rowDimensionValues: [],
              totals: [],
              data: [],
            };

            return axisSet(rowMeasureValuesEmpty);
          }

          const subRowDimensionValues =
            subRowDimensions?.data?.[anchorDimension] || [];

          // Detect if more data is available (we got limit+1 results)
          const hasMore = subRowDimensionValues.length > effectiveLimit;

          // Trim to effective limit (remove the +1 row we fetched for detection)
          const trimmedValues = hasMore
            ? subRowDimensionValues.slice(0, effectiveLimit)
            : subRowDimensionValues;

          // Also trim the totals to match the trimmed values
          const allSubRowDimensionTotals =
            subRowDimensions?.totals?.[anchorDimension] || [];
          const subRowDimensionTotals = hasMore
            ? allSubRowDimensionTotals.slice(0, effectiveLimit)
            : allSubRowDimensionTotals;

          const subRowAxesQueryForMeasureTotals = getAxisQueryForMeasureTotals(
            ctx,
            config,
            isMeasureSortAccessor,
            sortAccessor,
            anchorDimension,
            trimmedValues,
            timeRangeRow,
            subTableMergedFilters,
          );

          let subTableQuery:
            | Readable<null>
            | CreateQueryResult<
                V1MetricsViewAggregationResponse,
                ConnectError
              > = readable(null);

          if (config.colDimensionNames.length) {
            subTableQuery = createSubTableCellQuery(
              ctx,
              config,
              anchorDimension,
              columnDimensionAxesData,
              totalsRow,
              trimmedValues,
              subTableMergedFilters,
              rowNestTimeFilters,
            );
          }

          return derived(
            [subRowAxesQueryForMeasureTotals, subTableQuery],
            ([subRowTotals, subTableData]) => {
              const subTableQueryError = getErrorFromResponses([subTableData]);
              if (subTableQueryError.length || subRowTotals?.error?.length) {
                const allErrors = subTableQueryError.concat(
                  subRowTotals?.error || [],
                );
                return getExpandedErrorState(allErrors, expandIndex);
              }
              if (subRowTotals?.isFetching) {
                const rowMeasureValueWithoutSubTable: ExpandedRowMeasureValues =
                  {
                    isFetching: true,
                    expandIndex,
                    rowDimensionValues: trimmedValues,
                    totals: [],
                    data: subTableData?.data?.data || [],
                    hasMore,
                    effectiveLimit,
                  };
                return rowMeasureValueWithoutSubTable;
              }

              const mergedRowTotals = mergeRowTotalsInOrder(
                trimmedValues,
                subRowDimensionTotals,
                subRowTotals?.data?.[anchorDimension] || [],
                subRowTotals?.totals?.[anchorDimension] || [],
              );

              if (!subTableData) {
                const rowMeasureValueWithoutSubTable: ExpandedRowMeasureValues =
                  {
                    isFetching: false,
                    expandIndex,
                    rowDimensionValues: trimmedValues,
                    totals: mergedRowTotals,
                    data: mergedRowTotals,
                    hasMore,
                    effectiveLimit,
                  };
                return rowMeasureValueWithoutSubTable;
              }

              if (subTableData?.isFetching) {
                const rowMeasureValueWithTotals: ExpandedRowMeasureValues = {
                  isFetching: true,
                  expandIndex,
                  rowDimensionValues: trimmedValues,
                  totals: mergedRowTotals,
                  data: subTableData?.data?.data || [],
                  hasMore,
                  effectiveLimit,
                };

                return rowMeasureValueWithTotals;
              }

              const rowMeasureValue: ExpandedRowMeasureValues = {
                isFetching: false,
                expandIndex,
                rowDimensionValues: trimmedValues,
                totals: mergedRowTotals,
                data: subTableData?.data?.data || [],
                hasMore,
                effectiveLimit,
              };
              return rowMeasureValue;
            },
          ).subscribe(axisSet);
        },
      );
    }) ?? [],
    (combos: ExpandedRowMeasureValues[]) => {
      return combos;
    },
  );
}

/***
 * For each expanded row, add the sub table data to the pivot table
 * data at the correct position.
 *
 * Note: Since the nested dimension values are present in the outermost
 * dimension's column, their accessor is the same as the anchor dimension.
 * Therefore, we change the key of the nested dimension to the anchor.
 */
export function addExpandedDataToPivot(
  config: PivotDataStoreConfig,
  tableData: PivotDataRow[],
  rowDimensions: string[],
  columnDimensionAxes: Record<string, string[]>,
  expandedRowMeasureValues: ExpandedRowMeasureValues[],
): PivotDataRow[] {
  const pivotData = tableData;
  const numRowDimensions = rowDimensions.length;

  // Fill parents before their children: a parent's sub rows are replaced
  // wholesale, which would drop children filled in earlier. The order of
  // `expanded` can't be relied on, e.g. after collapsing and re-expanding a
  // parent its key comes after its children's.
  const parentsFirst = [...expandedRowMeasureValues].sort(
    (a, b) => expandKeyDepth(a.expandIndex) - expandKeyDepth(b.expandIndex),
  );

  parentsFirst.forEach((expandedRowData) => {
    const rowValues = expandedRowData.rowDimensionValues;

    if (rowValues.length === 0) return;
    const segments = expandKeySegments(expandedRowData.expandIndex);

    const rows = getRowsForExpandedKey(
      pivotData,
      rowDimensions,
      expandedRowData.expandIndex,
    );
    if (rows.length !== segments.length) return;
    const node = rows.at(-1);

    // Update the node's subRows in place
    if (node && node.subRows) {
      const anchorDimension = rowDimensions[segments.length];

      let skeletonSubTable: PivotDataRow[] = [
        { [anchorDimension]: LOADING_CELL },
      ];
      if (expandedRowData?.totals?.length) {
        skeletonSubTable = expandedRowData?.totals as PivotDataRow[];
      }

      let subTableData = skeletonSubTable;
      if (expandedRowData?.data?.length) {
        subTableData = reduceTableCellDataIntoRows(
          config,
          anchorDimension,
          rowValues,
          columnDimensionAxes,
          skeletonSubTable,
          expandedRowData?.data,
          true,
        );
      }

      const mappedSubRows =
        subTableData?.map((row) => {
          const newRow = {
            ...row,
            [rowDimensions[0]]: row[anchorDimension],
          };

          /**
           * Add sub rows to the new row if number of row dimensions
           * is greater than number of nest levels expanded except
           * for the last level
           */
          if (numRowDimensions - 1 > segments.length) {
            newRow.subRows = [{ [rowDimensions[0]]: LOADING_CELL }];
          }
          return newRow;
        }) || [];

      // Add a "Show more" row if more data is available. Past the last
      // limit step the row only notes that the rest is left out.
      if (expandedRowData.hasMore && expandedRowData.effectiveLimit) {
        mappedSubRows.push({
          [rowDimensions[0]]: SHOW_MORE_BUTTON,
          __currentLimit: expandedRowData.effectiveLimit,
        } as PivotDataRow);
      }

      node.subRows = mappedSubRows;
    }
  });
  return pivotData;
}

export function getExpandedQueryErrors(
  expandedRowMeasureValues: ExpandedRowMeasureValues[],
): PivotQueryError[] {
  return expandedRowMeasureValues
    .flatMap((expandedRow) => expandedRow.error || [])
    .filter((error) => error !== undefined);
}
