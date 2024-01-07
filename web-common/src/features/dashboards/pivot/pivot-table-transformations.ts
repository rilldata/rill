import { createIndexMap, getColumnDefForPivot } from "./pivot-utils";

/**
 * Create a barebone table with row and column headers.
 * This is used to render the table skeleton before cell data is fetched.
 */
export function createTableWithAxes(
  config,
  columnDimensionAxes,
  rowDimensionAxes
) {
  const columnDef = getColumnDefForPivot(config, columnDimensionAxes);

  let data = [];

  if (
    config.rowDimensionNames.length &&
    Object.keys(rowDimensionAxes)?.length
  ) {
    const anchorDimensionName = config.rowDimensionNames[0];
    data = rowDimensionAxes?.[anchorDimensionName]?.map((value) => {
      return {
        [anchorDimensionName]: value,
      };
    });
  }

  return {
    data,
    columnDef,
  };
}

/**
 * Create a nested accessor for a cell in the table.
 * This is used to map the cell data to the table data.
 *
 * Column names are converted to c0, c1, c2, etc.
 * Column values are converted to v0, v1, v2, etc.
 * Measure names are converted to m0, m1, m2, etc.
 */
export function getAccessorForCell(
  colDimensionNames,
  colValuesIndexMaps,
  numMeasures,
  cell
) {
  const nestedColumnValueAccessor = colDimensionNames
    .map((colName, i) => {
      let accessor = `c${i}`;

      const colValue = cell[colName];
      const colValueIndex = colValuesIndexMaps[i].get(colValue);
      accessor += `v${colValueIndex}`;

      return accessor;
    })
    .join("_");

  return Array(numMeasures)
    .fill(null)
    .map((_, i) => `${nestedColumnValueAccessor}m${i}`);
}

/***
 *
 * The table expects pivoted data for a single row to be mapped to a single object
 * This functions reduces the API response from
 *
 * +-------------------------------+    +------------------------------------------------+
 * | {                             |    | {                                              |
 * |  "row_dimension: value,       |    |   "row_dimension": value,                      |
 * |  "column_dimension: value_1,  | => |   "column_dimension_value_1": measure_value_1, |
 * |  "measure": measure_value_1   |    |   "column_dimension_value_2": measure_value_2  |
 * | },                            |    | }                                              |
 * | {                             |    +------------------------------------------------+
 * |  "row_dimension: value,       |    Transformed to a Tanstack table readable format
 * |  "column_dimension: value_2,  |
 * |  "measure": measure_value_2   |
 * | }                             |
 * +-------------------------------+
 */

export function reduceTableCellDataIntoRows(
  config,
  columnDimensionAxes,
  rowDimensionAxes,
  tableData,
  cellData
) {
  const rowDimensionName = config.rowDimensionNames[0];
  const colDimensionNames = config.colDimensionNames;

  // For simple tables, return the cell data as is
  if (!rowDimensionName || !colDimensionNames.length) {
    return cellData;
  }

  /**
   * Create a map of row dimension values to their index in the row dimension axes.
   * This way we can apply the sort order on the row dimension axes and sort the cell
   * data using the dimension values. For O(n) pass on the cells (n = number of cells),
   * we can reduce them optimally into row objects.
   */
  const rowDimensionIndexMap = createIndexMap(
    rowDimensionAxes[rowDimensionName]
  );

  const colValuesIndexMaps = colDimensionNames.map((colDimensionName) =>
    createIndexMap(columnDimensionAxes[colDimensionName])
  );

  cellData.forEach((cell) => {
    const rowDimensionValue = cell[rowDimensionName];
    const rowIndex = rowDimensionIndexMap.get(rowDimensionValue);
    const row = tableData[rowIndex];

    const accessors = getAccessorForCell(
      colDimensionNames,
      colValuesIndexMaps,
      config.measureNames.length,
      cell
    );

    if (row) {
      accessors.forEach((accessor, i) => {
        row[accessor] = cell[config.measureNames[i]];
      });
    }
  });

  return tableData;
}
