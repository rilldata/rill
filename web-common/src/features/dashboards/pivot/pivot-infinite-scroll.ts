import type {
  PivotDataRow,
  PivotDataStoreConfig,
  TimeFilters,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  extractNumbers,
  getTimeFilterSpanningValues,
  isTimeDimension,
  sortAcessors,
} from "./pivot-utils";

/**
 * We need to make sure that when we update these limits, we also update the
 * limit of the query in `pivot-data-store.ts`
 */
export const NUM_COLUMNS_PER_PAGE = 50;
export const NUM_ROWS_PER_PAGE = 50;

function getColumnsPerPage(config: PivotDataStoreConfig) {
  return config.enableComparison
    ? NUM_COLUMNS_PER_PAGE * 3
    : NUM_COLUMNS_PER_PAGE;
}

function getSortedColumnKeys(
  config: PivotDataStoreConfig,
  totalsRow: PivotDataRow,
) {
  const { measureNames, rowDimensionNames } = config;

  const allColumnKeys = totalsRow ? Object.keys(totalsRow) : [];
  const colHeaderKeys = allColumnKeys.filter(
    (key) => !(measureNames.includes(key) || rowDimensionNames[0] === key),
  );
  return sortAcessors(colHeaderKeys);
}

/**
 * Parse a column accessor such as `c0v3_c1v0m2` into the index of each
 * column dimension and of its value.
 */
function getColumnValueIndices(accessor: string) {
  // Strip the measure string from the accessor
  const [accessorWithoutMeasure] = accessor.split("m");
  return accessorWithoutMeasure.split("_").map((part) => extractNumbers(part));
}

/**
 * Slice column axes data based on page
 * number. This is used for column definition in pivot table.
 * Also counts the leaf columns the sliced axes render,
 * out of the leaf columns the totals row has.
 */
export function sliceColumnAxesDataForDef(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  totalsRow: PivotDataRow,
): {
  axes: Record<string, string[]>;
  shownColumns: number;
  totalColumns: number;
} {
  const { colDimensionNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;
  if (!colDimensionNames.length) {
    return { axes: colDimensionAxes, shownColumns: 0, totalColumns: 0 };
  }

  const totalColumnsToBeDisplayed =
    getColumnsPerPage(config) * colDimensionPageNumber;

  const maxIndexVisible: Record<string, number> = {};

  const sortedColumnKeys = getSortedColumnKeys(config, totalsRow);

  const columnKeysForPage = sortedColumnKeys.slice(
    0,
    totalColumnsToBeDisplayed,
  );

  columnKeysForPage.forEach((accessor) => {
    getColumnValueIndices(accessor).forEach(({ c, v }) => {
      const columnDimensionName = colDimensionNames[c];
      maxIndexVisible[columnDimensionName] = Math.max(
        maxIndexVisible[columnDimensionName] || 0,
        v + 1,
      );
    });
  });

  const slicedAxesData: Record<string, string[]> = {};

  Object.keys(maxIndexVisible).forEach((dimensionName) => {
    if (maxIndexVisible[dimensionName] > 0) {
      slicedAxesData[dimensionName] = colDimensionAxes[dimensionName].slice(
        0,
        maxIndexVisible[dimensionName],
      );
    }
  });

  // The column definition renders every combination of the sliced values
  // that the totals row has, which can be a few more than the page holds.
  const shownColumns = sortedColumnKeys.filter((accessor) =>
    getColumnValueIndices(accessor).every(
      ({ c, v }) => v < (slicedAxesData[colDimensionNames[c]]?.length ?? 0),
    ),
  ).length;

  return {
    axes: slicedAxesData,
    shownColumns,
    totalColumns: sortedColumnKeys.length,
  };
}

/**
 * Slice the column dimension values to the right limit using column
 * page number and page size
 */
export function getColumnFiltersForPage(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  totalsRow: PivotDataRow,
) {
  const { measureNames, colDimensionNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;
  const totalColumnsToBeDisplayed = getColumnsPerPage(config);

  if (!colDimensionNames.length || measureNames.length == 0)
    return { filters: [], timeFilters: [] };

  const pageStartIndex =
    totalColumnsToBeDisplayed * (colDimensionPageNumber - 1);

  const sortedColumnKeys = getSortedColumnKeys(config, totalsRow);

  const columnKeysForPage = sortedColumnKeys.slice(
    pageStartIndex,
    pageStartIndex + totalColumnsToBeDisplayed,
  );

  const minIndexVisible: Record<string, number> = {};
  const maxIndexVisible: Record<string, number> = {};

  columnKeysForPage.forEach((accessor) => {
    getColumnValueIndices(accessor).forEach(({ c, v }) => {
      const dimension = colDimensionNames[c];
      maxIndexVisible[dimension] = Math.max(
        maxIndexVisible[dimension] || 0,
        v + 1,
      );
      minIndexVisible[dimension] = Math.min(
        minIndexVisible[dimension] ?? Number.MAX_SAFE_INTEGER,
        v,
      );
    });
  });

  const slicedAxesData: Record<string, string[]> = {};

  Object.keys(minIndexVisible).forEach((dimension) => {
    slicedAxesData[dimension] = colDimensionAxes[dimension].slice(
      minIndexVisible[dimension],
      maxIndexVisible[dimension],
    );
  });

  const timeFilters: TimeFilters[] = [];
  const filters = colDimensionNames
    .filter((dimension) => {
      if (isTimeDimension(dimension, config.time.timeDimension)) {
        timeFilters.push(
          getTimeFilterSpanningValues(dimension, slicedAxesData[dimension]),
        );
        return false;
      }
      return true;
    })
    .map((dimension) =>
      createInExpression(dimension, slicedAxesData[dimension]),
    );

  return { filters, timeFilters };
}
