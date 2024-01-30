import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  V1Expression,
  V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import { Readable, derived, writable } from "svelte/store";
import {
  createPivotAggregationRowQuery,
  getAxisForDimensions,
} from "./pivot-queries";
import { reduceTableCellDataIntoRows } from "./pivot-table-transformations";
import {
  getFilterForPivotTable,
  getSortForAccessor,
  getTimeForQuery,
} from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig, TimeFilters } from "./types";

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
function getValuesForExpandedKey(
  tableData: PivotDataRow[],
  rowDimensions: string[],
  key: string,
) {
  const indices = key.split(".").map((index) => parseInt(index, 10));

  // The first row is always the totals row for the expanded context
  indices[0] = indices[0] - 1;

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
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  anchorDimension: string,
  columnDimensionAxesData: Record<string, string[]> | undefined,
  rowNestFilters: V1Expression,
  timeRange: TimeRangeString | undefined,
) {
  const allDimensions = config.colDimensionNames.concat([anchorDimension]);

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

  const filterForSubTable = getFilterForPivotTable(
    config,
    columnDimensionAxesData,
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
    config.measureNames,
    dimensionBody,
    filterForSubTable,
    sortBy,
    "10000",
    "0",
    timeRange,
  );
}

interface ExpandedRowMeasureValues {
  isFetching: boolean;
  expandIndex: string;
  rowDimensionValues: string[];
  data: V1MetricsViewAggregationResponseDataItem[];
  totals: V1MetricsViewAggregationResponseDataItem[];
}

/**
 * For each expanded row, create a query for the sub table
 * and return the query result along with the expanded row index
 * and the row dimension values
 */
export function queryExpandedRowMeasureValues(
  ctx: StateManagers,
  config: PivotDataStoreConfig,
  tableData: PivotDataRow[],
  columnDimensionAxesData: Record<string, string[]> | undefined,
): Readable<ExpandedRowMeasureValues[] | null> {
  const { rowDimensionNames } = config;
  const expanded = config.pivot.expanded;
  if (!tableData || Object.keys(expanded).length == 0) return writable(null);

  return derived(
    Object.keys(expanded)?.map((expandIndex) => {
      const nestLevel = expandIndex?.split(".")?.length;
      const anchorDimension = rowDimensionNames[nestLevel];
      const values = getValuesForExpandedKey(
        tableData,
        rowDimensionNames,
        expandIndex,
      );

      if (
        !anchorDimension ||
        !values.length ||
        values.some((v) => v === undefined || v === "LOADING_CELL")
      )
        return writable({
          isFetching: true,
          expandIndex,
          rowDimensionValues: [],
          data: [],
          totals: [],
        });

      const timeFilters: TimeFilters[] = [];
      // TODO: handle for already existing filters
      const rowNestFilters = values
        .map((value, index) =>
          createInExpression(rowDimensionNames[index], [value]),
        )
        .filter((f) => {
          // We map first and filter later to ensure that dimensions are in order
          if (f.cond?.exprs?.[0].ident === config.time.timeDimension) {
            timeFilters.push({
              timeStart: f.cond?.exprs?.[1].val as string,
              interval: config.time.interval,
            });
            return false;
          } else return true;
        });

      const filterForRowDimensionAxes = createAndExpression(rowNestFilters);

      const { sortPivotBy } = getSortForAccessor(
        anchorDimension,
        config,
        columnDimensionAxesData,
      );

      const timeRange: TimeRangeString = getTimeForQuery(
        config.time,
        timeFilters,
      );

      // const mergedFilter = mergeFilters(
      //   createAndExpression(rowNestFilters),
      //   config.whereFilter,
      // );

      return derived(
        [
          writable(expandIndex),
          getAxisForDimensions(
            ctx,
            config,
            [anchorDimension],
            filterForRowDimensionAxes,
            sortPivotBy,
            timeRange,
          ),
          createSubTableCellQuery(
            ctx,
            config,
            anchorDimension,
            columnDimensionAxesData,
            createAndExpression(rowNestFilters),
            timeRange,
          ),
        ],
        ([expandIndex, subRowDimensions, subTableData]) => {
          return {
            isFetching: subTableData?.isFetching,
            expandIndex,
            rowDimensionValues: subRowDimensions?.data?.[anchorDimension] || [],
            totals: subRowDimensions?.totals?.[anchorDimension] || [],
            data: subTableData?.data?.data || [],
          };
        },
      );
    }),
    (combos) => {
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
    const indices = expandedRowData.expandIndex
      .split(".")
      .map((index) => parseInt(index, 10));

    // The first row is always the totals row for the expanded context
    indices[0] = indices[0] - 1;

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
      const rowValues = expandedRowData.rowDimensionValues;

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
