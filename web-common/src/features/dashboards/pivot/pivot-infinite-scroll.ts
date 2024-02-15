import { extractNumbers } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

const NUM_COLUMNS_PER_PAGE = 50;

type ColumnNode = {
  value: string;
  depth: number;
};

function getTotalNodes(tree: Array<Array<string>>): number {
  return tree.reduce((acc, level) => acc * level.length, 1);
}

function getColumnPage(
  tree: Array<Array<string>>,
  pageNumber: number,
  pageSize: number,
): Record<number, Set<string>> {
  const startIndex = (pageNumber - 1) * pageSize;
  const totalNodes = getTotalNodes(tree);

  if (startIndex >= totalNodes || startIndex < 0) {
    return []; // Page number out of range
  }

  const result: ColumnNode[] = [];
  let currentIndex = 0;

  function dfs(level: number, path: ColumnNode[]) {
    if (level === tree.length) {
      if (currentIndex >= startIndex && currentIndex < startIndex + pageSize) {
        result.push(...path);
      }
      currentIndex++;
      return;
    }
    for (const child of tree[level]) {
      if (currentIndex >= startIndex + pageSize) {
        break; // Stop processing once the page is filled
      }
      dfs(level + 1, [...path, { value: child, depth: level }]);
    }
  }

  dfs(0, []);

  const groups: Record<number, Set<string>> = {};

  result.forEach(({ value, depth }) => {
    groups[depth] = groups[depth] || new Set();
    groups[depth].add(value);
  });

  return groups;
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
  const { rowDimensionNames, colDimensionNames, measureNames } = config;
  const colDimensionPageNumber = config.pivot.columnPage;
  if (!colDimensionNames.length) return colDimensionAxes;

  const totalColumnsToBeDisplayed =
    NUM_COLUMNS_PER_PAGE * colDimensionPageNumber;

  const maxValueVisible: Record<string, number> = {};

  let columnKeys = totalsRow ? Object.keys(totalsRow) : [];
  columnKeys = columnKeys
    .filter(
      (key) => !(measureNames.includes(key) || rowDimensionNames[0] === key),
    )
    .sort()
    .slice(0, totalColumnsToBeDisplayed);

  console.log("columnKeys", columnKeys);

  columnKeys.forEach((accessor) => {
    // Strip the measure string from the accessor
    const [accessorWithoutMeasure] = accessor.split("m");
    accessorWithoutMeasure.split("_").forEach((part) => {
      const { c, v } = extractNumbers(part);
      const columnDimensionName = colDimensionNames[c];
      maxValueVisible[columnDimensionName] = Math.max(
        maxValueVisible[columnDimensionName] || 0,
        v + 1,
      );
    });
  });

  const slicedAxesData: Record<string, string[]> = {};

  Object.keys(maxValueVisible).forEach((dimensionName) => {
    if (maxValueVisible[dimensionName] > 0) {
      slicedAxesData[dimensionName] = colDimensionAxes[dimensionName].slice(
        0,
        maxValueVisible[dimensionName],
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
  colDimensionNames: string[],
  colDimensionAxes: Record<string, string[]> = {},
  colDimensionPageNumber: number,
  numMeasures: number,
): V1Expression[] {
  if (!colDimensionNames.length || numMeasures == 0) return [];

  const effectiveColumnsPerPage = Math.floor(
    NUM_COLUMNS_PER_PAGE / numMeasures,
  );

  const colDimensionValues = colDimensionNames.map((colDimensionName) => {
    return colDimensionAxes[colDimensionName];
  });

  const pageGroups = getColumnPage(
    colDimensionValues,
    colDimensionPageNumber,
    effectiveColumnsPerPage,
  );

  return Object.entries(pageGroups).map(([colDimensionId, values]) =>
    createInExpression(
      colDimensionNames[parseInt(colDimensionId)],
      Array.from(values),
    ),
  );
}
