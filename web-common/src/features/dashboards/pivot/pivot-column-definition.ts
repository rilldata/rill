import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  addZoneOffset,
  removeLocalTimezoneOffset,
} from "@rilldata/web-common/lib/time/timezone";
import type { ColumnDef } from "@tanstack/svelte-table";
import { timeFormat } from "d3-time-format";
import PivotExpandableCell from "./PivotExpandableCell.svelte";
import PivotMeasureCell from "./PivotMeasureCell.svelte";
import {
  cellComponent,
  createIndexMap,
  getAccessorForCell,
} from "./pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
  PivotTimeConfig,
} from "./types";

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  config: PivotDataStoreConfig,
  colDimensions: { label: string; name: string }[],
  headers: Record<string, string[]>,
  leafData: ColumnDef<PivotDataRow>[],
): ColumnDef<PivotDataRow>[] {
  const dimensionNames = config.colDimensionNames;
  const timeConfig = config.time;

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
    return headerValues?.map((value) => {
      let displayValue = value;
      if (timeConfig?.timeDimension === dimensionNames?.[level]) {
        const timeGrain = timeConfig?.interval;
        const dt = addZoneOffset(
          removeLocalTimezoneOffset(new Date(value)),
          timeConfig?.timeZone,
        );
        const timeFormatter = timeFormat(
          timeGrain ? TIME_GRAIN[timeGrain].d3format : "%H:%M",
        ) as (d: Date) => string;

        displayValue = timeFormatter(dt);
      } else if (displayValue === null) displayValue = "null";
      return {
        header: displayValue,
        columns: createNestedColumns(level + 1, {
          ...colValuePair,
          [dimensionNames[level]]: value,
        }),
      };
    });
  }

  // Construct column def for Row Totals
  let rowTotalsColumns: ColumnDef<PivotDataRow>[] = [];
  if (config.rowDimensionNames.length && config.colDimensionNames.length) {
    rowTotalsColumns = colDimensions.reverse().reduce((acc, dimension) => {
      const { label, name } = dimension;

      const headColumn = {
        header: label || name,
        columns: acc,
      };

      return [headColumn];
    }, leafData);
  }

  // Start the recursion
  const nestedColumns = createNestedColumns(0, {});

  return [...rowTotalsColumns, ...nestedColumns];
}

/**
 * Get formatted value for row dimension values. Format
 * time dimension values if present.
 */
function formatRowDimensionValue(
  value: string,
  depth: number,
  timeConfig: PivotTimeConfig,
  rowDimensionNames: string[],
) {
  const dimension = rowDimensionNames?.[depth];
  if (dimension === timeConfig?.timeDimension) {
    if (value === "Total") return "Total";
    const timeGrain = timeConfig?.interval;
    const dt = addZoneOffset(
      removeLocalTimezoneOffset(new Date(value)),
      timeConfig?.timeZone,
    );
    const timeFormatter = timeFormat(
      timeGrain ? TIME_GRAIN[timeGrain]?.d3format : "%H:%M",
    ) as (d: Date) => string;

    return timeFormatter(dt);
  }
  return value;
}

/**
 * Create column definitions object for pivot table
 * as required by Tanstack Table
 */
export function getColumnDefForPivot(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> | undefined,
) {
  const IsNested = true;

  const { measureNames, rowDimensionNames, colDimensionNames } = config;

  const measures = measureNames.map((m) => {
    const measure = config.allMeasures.find((measure) => measure.name === m);

    if (!measure) {
      throw new Error(`Measure ${m} not found in config.allMeasures`);
    }

    return {
      label: measure?.label || m,
      formatter: createMeasureValueFormatter<null | undefined>(measure),
      name: m,
    };
  });

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
            value: formatRowDimensionValue(
              getValue() as string,
              row.depth,
              config.time,
              rowDimensionNames,
            ),
            row,
          }),
      };
    });

  const leafColumns: ColumnDef<PivotDataRow>[] = measures.map((m) => {
    return {
      accessorKey: m.name,
      header: m.label || m.name,
      cell: (info) => {
        const value = m.formatter(info.getValue() as number | null | undefined);

        if (value == null) return cellComponent(PivotMeasureCell, {});
        return value;
      },
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    config,
    colDimensions,
    columnDimensionAxes || {},
    leafColumns,
  );

  return [...rowDefinitions, ...groupedColDef];
}
