import type { ColumnDef } from "@tanstack/svelte-table";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import {
  cellComponent,
  createIndexMap,
  getAccessorForCell,
} from "./pivot-utils";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  dimensionNames: string[],
  headers: Record<string, string[]>,
  leafData: ColumnDef<PivotDataRow>[],
): ColumnDef<PivotDataRow>[] {
  const colValuesIndexMaps = dimensionNames?.map((colDimension) =>
    createIndexMap(headers[colDimension]),
  );

  const levels = dimensionNames.length;
  // Recursive function to create nested columns
  function createNestedColumns(
    level: number,
    colValuePair: { [key: string]: string },
  ): ColumnDef<PivotDataRow>[] {
    if (level === levels) {
      const accessors = getAccessorForCell(
        dimensionNames,
        colValuesIndexMaps,
        leafData.length,
        colValuePair,
      );

      // Base case: return leaf columns
      return leafData.map((leaf, i) => ({
        ...leaf,
        // Change accessor key to match the nested column structure
        accessorKey: accessors[i],
      }));
    }

    // Recursive case: create nested headers
    const headerValues = headers[dimensionNames?.[level]];
    return headerValues?.map((value) => ({
      header: value,
      columns: createNestedColumns(level + 1, {
        ...colValuePair,
        [dimensionNames[level]]: value,
      }),
    }));
  }

  // Start the recursion
  return createNestedColumns(0, {});
}

export function getColumnDefForPivot(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> | undefined,
) {
  const IsNested = true;

  const { measureNames, rowDimensionNames, colDimensionNames } = config;

  const measures = measureNames.map((m) => ({
    label: config.allMeasures.find((measure) => measure.name === m)?.label || m,
    name: m,
  }));

  const rowDimensions = rowDimensionNames.map((d) => ({
    label:
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d,
    name: d,
  }));
  const colDimensions = colDimensionNames.map((d) => ({
    label:
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d,
    name: d,
  }));

  let rowDimensionsForColumnDef = rowDimensions;
  let nestedLabel: string;
  if (IsNested) {
    rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
    nestedLabel = rowDimensions.map((d) => d.label || d.name).join(" > ");
  }
  const rowDefinitions: ColumnDef<PivotDataRow>[] =
    rowDimensionsForColumnDef.map((d) => {
      return {
        accessorKey: d.name,
        header: nestedLabel,
        cell: ({ row, getValue }) =>
          cellComponent(PivotExpandableCell, {
            value: getValue(),
            row,
          }),
      };
    });

  const leafColumns: ColumnDef<PivotDataRow>[] = measures.map((m) => {
    return {
      accessorKey: m.name,
      header: m.label || m.name,
      cell: (info) => info.getValue(),
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    colDimensions.map((d) => d.name) || [],
    columnDimensionAxes || {},
    leafColumns,
  );

  return [...rowDefinitions, ...groupedColDef];
}
