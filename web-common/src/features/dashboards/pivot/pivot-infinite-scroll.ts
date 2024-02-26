import type {
  PivotDataRow,
  PivotDataStoreConfig,
  TimeFilters,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  extractNumbers,
  getTimeGrainFromDimension,
  isTimeDimension,
  sortAcessors,
} from "./pivot-utils";

const NUM_COLUMNS_PER_PAGE = 50;

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
 * Slice column axes data based on page
 * number. This is used for column definition in pivot table.
 */
export function sliceColumnAxesDataForDef(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  totalsRow: PivotDataRow,
) {
  const { colDimensionNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;
  if (!colDimensionNames.length) return colDimensionAxes;

  const totalColumnsToBeDisplayed =
    NUM_COLUMNS_PER_PAGE * colDimensionPageNumber;

  const maxIndexVisible: Record<string, number> = {};

  const sortedColumnKeys = getSortedColumnKeys(config, totalsRow);

  const columnKeysForPage = sortedColumnKeys.slice(
    0,
    totalColumnsToBeDisplayed,
  );

  columnKeysForPage.forEach((accessor) => {
    // Strip the measure string from the accessor
    const [accessorWithoutMeasure] = accessor.split("m");
    accessorWithoutMeasure.split("_").forEach((part) => {
      const { c, v } = extractNumbers(part);
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

  return slicedAxesData;
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

  if (!colDimensionNames.length || measureNames.length == 0)
    return { filters: [], timeFilters: [] };

  const pageStartIndex = NUM_COLUMNS_PER_PAGE * (colDimensionPageNumber - 1);

  const sortedColumnKeys = getSortedColumnKeys(config, totalsRow);

  const columnKeysForPage = sortedColumnKeys.slice(
    pageStartIndex,
    pageStartIndex + NUM_COLUMNS_PER_PAGE,
  );

  const minIndexVisible: Record<string, number> = {};
  const maxIndexVisible: Record<string, number> = {};

  columnKeysForPage.forEach((accessor) => {
    // Strip the measure string from the accessor
    const [accessorWithoutMeasure] = accessor.split("m");
    accessorWithoutMeasure.split("_").forEach((part) => {
      const { c, v } = extractNumbers(part);
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
        const dates = slicedAxesData[dimension].map((d) =>
          new Date(d).getTime(),
        );
        const timeStart = new Date(Math.min(...dates)).toISOString();
        const timeEnd = new Date(Math.max(...dates)).toISOString();
        const interval = getTimeGrainFromDimension(dimension);
        timeFilters.push({ timeStart, timeEnd, interval });
        return false;
      }
      return true;
    })
    .map((dimension) =>
      createInExpression(dimension, slicedAxesData[dimension]),
    );

  return { filters, timeFilters };
}
