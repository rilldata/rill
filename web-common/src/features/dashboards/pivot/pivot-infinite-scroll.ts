import { extractNumbers } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

const NUM_COLUMNS_PER_PAGE = 50;

/**
 * Slice column axes data based on page
 * number. This is used for column definition in pivot table.
 */
export function sliceColumnAxesDataForDef(
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
  totalsRow: PivotDataRow,
) {
  const { rowDimensionNames, colDimensionNames, measureNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;
  if (!colDimensionNames.length) return colDimensionAxes;

  const totalColumnsToBeDisplayed =
    NUM_COLUMNS_PER_PAGE * colDimensionPageNumber;

  const maxIndexVisible: Record<string, number> = {};

  let columnKeys = totalsRow ? Object.keys(totalsRow) : [];
  columnKeys = columnKeys
    .filter(
      (key) => !(measureNames.includes(key) || rowDimensionNames[0] === key),
    )
    .sort()
    .slice(0, totalColumnsToBeDisplayed);

  columnKeys.forEach((accessor) => {
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
): V1Expression[] {
  const { measureNames, colDimensionNames, rowDimensionNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;

  if (!colDimensionNames.length || measureNames.length == 0) return [];

  const pageStartIndex = NUM_COLUMNS_PER_PAGE * (colDimensionPageNumber - 1);

  const allColumnKeys = totalsRow ? Object.keys(totalsRow) : [];
  const columnKeysForPage = allColumnKeys
    .filter(
      (key) => !(measureNames.includes(key) || rowDimensionNames[0] === key),
    )
    .sort()
    .slice(pageStartIndex, pageStartIndex + NUM_COLUMNS_PER_PAGE);

  const minIndexVisible: Record<string, number> = {};
  const maxIndexVisible: Record<string, number> = {};

  console.log(columnKeysForPage);

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

  console.log(slicedAxesData);

  return Object.entries(slicedAxesData).map(([dimension, values]) =>
    createInExpression(dimension, Array.from(values)),
  );
}
