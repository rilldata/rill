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
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
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
 * Extracts and organizes dimension values from a nested array structure
 * based on a given dimensions and an expanded key.
 *
 * This function iterates over a key in the `expanded` object, which
 * indicates whether a particular path in the nested array is expanded.
 * For each expanded path, it navigates through the table data
 * following the path defined by the key (split into indices) and extracts
 * the dimension values at each level.
 */
export function getValuesForExpandedKey(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  key: string,
  hasTotalsRow = true,
): string[] {
  const indices = key.split(".").map((index) => parseInt(index, 10));

  if (hasTotalsRow) {
    // The first row is always the totals row for the expanded context with measures
    indices[0] = indices[0] - 1;
  }

  // Retrieve the value from the nested array
  let currentValue: PivotDataRow[] | undefined = tableData;
  const dimensionValues: string[] = [];

  indices.forEach((index, i) => {
    if (!currentValue?.[index]) {
      return;
    }
    dimensionValues.push(currentValue[index]?.[rowDimensions[i]] as string);
    currentValue = currentValue[index]?.subRows;
  });
  return dimensionValues;
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
      const nestLevel = expandIndex?.split(".")?.length;

      if (nestLevel >= rowDimensionNames.length)
        return readable({
          isFetching: false,
          expandIndex,
          rowDimensionValues: [],
          data: [],
          totals: [],
        });
      const anchorDimension = rowDimensionNames[nestLevel];
      const values = getValuesForExpandedKey(
        tableData,
        rowDimensionNames,
        expandIndex,
        numMeasures > 0,
      );

      if (
        !anchorDimension ||
        !values.length ||
        values.some((v) => v === undefined || v === "LOADING_CELL")
      )
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
          const subRowDimensionTotals =
            subRowDimensions?.totals?.[anchorDimension] || [];

          const subRowAxesQueryForMeasureTotals = getAxisQueryForMeasureTotals(
            ctx,
            config,
            isMeasureSortAccessor,
            sortAccessor,
            anchorDimension,
            subRowDimensionValues,
            timeRangeRow,
            subTableMergedFilters,
          );

          let subTableQuery:
            | Readable<null>
            | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
            readable(null);

          if (config.colDimensionNames.length) {
            subTableQuery = createSubTableCellQuery(
              ctx,
              config,
              anchorDimension,
              columnDimensionAxesData,
              totalsRow,
              subRowDimensionValues,
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
                    rowDimensionValues: subRowDimensionValues,
                    totals: [],
                    data: subTableData?.data?.data || [],
                  };
                return rowMeasureValueWithoutSubTable;
              }

              const mergedRowTotals = mergeRowTotalsInOrder(
                subRowDimensionValues,
                subRowDimensionTotals,
                subRowTotals?.data?.[anchorDimension] || [],
                subRowTotals?.totals?.[anchorDimension] || [],
              );

              if (!subTableData) {
                const rowMeasureValueWithoutSubTable: ExpandedRowMeasureValues =
                  {
                    isFetching: false,
                    expandIndex,
                    rowDimensionValues: subRowDimensionValues,
                    totals: mergedRowTotals,
                    data: mergedRowTotals,
                  };
                return rowMeasureValueWithoutSubTable;
              }

              if (subTableData?.isFetching) {
                const rowMeasureValueWithTotals: ExpandedRowMeasureValues = {
                  isFetching: true,
                  expandIndex,
                  rowDimensionValues: subRowDimensionValues,
                  totals: mergedRowTotals,
                  data: subTableData?.data?.data || [],
                };

                return rowMeasureValueWithTotals;
              }

              const rowMeasureValue: ExpandedRowMeasureValues = {
                isFetching: false,
                expandIndex,
                rowDimensionValues: subRowDimensionValues,
                totals: mergedRowTotals,
                data: subTableData?.data?.data || [],
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

  expandedRowMeasureValues.forEach((expandedRowData) => {
    const rowValues = expandedRowData.rowDimensionValues;

    if (rowValues.length === 0) return;
    const indices = expandedRowData.expandIndex
      .split(".")
      .map((index) => parseInt(index, 10));

    if (config.measureNames.length > 0) {
      // The first row is always the totals row for the expanded context with measures
      indices[0] = indices[0] - 1;
    }

    let parent: PivotDataRow[] = pivotData; // Keep a reference to the parent array
    let lastIdx = 0;

    // Traverse the data array to the right position
    for (let i = 0; i < indices.length; i++) {
      if (!parent[indices[i]]) break;
      if (i < indices.length - 1) {
        const subRows = parent[indices[i]].subRows;
        if (!subRows) break;
        parent = subRows;
      }
      lastIdx = indices[i];
    }

    // Update the specific array at the position
    if (parent[lastIdx] && parent[lastIdx].subRows) {
      const anchorDimension = rowDimensions[indices.length];

      let skeletonSubTable: PivotDataRow[] = [
        { [anchorDimension]: "LOADING_CELL" },
      ];
      if (expandedRowData?.totals?.length) {
        skeletonSubTable = expandedRowData?.totals;
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

      parent[lastIdx].subRows = subTableData?.map((row) => {
        const newRow = {
          ...row,
          [rowDimensions[0]]: row[anchorDimension],
        };

        /**
         * Add sub rows to the new row if number of row dimensions
         * is greater than number of nest levels expanded except
         * for the last level
         */
        if (numRowDimensions - 1 > indices.length) {
          newRow.subRows = [{ [rowDimensions[0]]: "LOADING_CELL" }];
        }
        return newRow;
      });
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
