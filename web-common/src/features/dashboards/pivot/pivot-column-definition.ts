import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  addZoneOffset,
  removeLocalTimezoneOffset,
} from "@rilldata/web-common/lib/time/timezone";
import type { ColumnDef } from "@tanstack/svelte-table";
import { timeFormat } from "d3-time-format";
import {
  createIndexMap,
  getAccessorForCell,
  getTimeGrainFromDimension,
  isTimeDimension,
} from "./pivot-utils";
import type {
  PivotDataRow,
  PivotDataStoreConfig,
  PivotTimeConfig,
} from "./types";

function sanitizeHeaderValue(value: unknown): string {
  if (value === "") return "\u00A0";
  if (typeof value === "string") return value;
  return String(value);
}

/***
 * Create nested and grouped column definitions for pivot table
 */
function createColumnDefinitionForDimensions(
  config: PivotDataStoreConfig,
  colDimensions: { label: string; name: string }[],
  headers: Record<string, string[]>,
  leafData: ColumnDef<PivotDataRow>[],
  totals: PivotDataRow,
): ColumnDef<PivotDataRow>[] {
  const dimensionNames = config.colDimensionNames;
  const timeConfig = config.time;

  const filterColumns = Boolean(dimensionNames.length);

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
        config.time.timeDimension,
        colValuesIndexMaps,
        leafData.length,
        colValuePair,
      );

      // Base case: return leaf columns
      const leafNodes = leafData.map((leaf, i) => ({
        ...leaf,
        // Change accessor key to match the nested column structure
        accessorKey: accessors[i],
      }));

      if (!filterColumns) {
        return leafNodes;
      }
      return leafNodes.filter((leaf) =>
        Object.keys(totals).includes(leaf.accessorKey),
      );
    }

    // Recursive case: create nested headers
    const headerValues = headers[dimensionNames?.[level]];
    return headerValues
      ?.map((value) => {
        let displayValue = value;
        if (
          isTimeDimension(dimensionNames?.[level], timeConfig?.timeDimension)
        ) {
          const timeGrain = getTimeGrainFromDimension(dimensionNames?.[level]);
          const dt = addZoneOffset(
            removeLocalTimezoneOffset(new Date(value)),
            timeConfig?.timeZone,
          );
          const timeFormatter = timeFormat(
            timeGrain ? TIME_GRAIN[timeGrain].d3format : "%H:%M",
          ) as (d: Date) => string;

          displayValue = timeFormatter(dt);
        }

        const nestedColumns = createNestedColumns(level + 1, {
          ...colValuePair,
          [dimensionNames[level]]: value,
        });

        return {
          header: sanitizeHeaderValue(displayValue),
          columns: nestedColumns,
        };
      })
      .filter((column) => column.columns.length > 0);
  }

  // Construct column def for Row Totals
  let rowTotalsColumns: ColumnDef<PivotDataRow>[] = [];
  if (config.rowDimensionNames.length && config.colDimensionNames.length) {
    rowTotalsColumns = colDimensions.reverse().reduce((acc, dimension) => {
      const { label, name } = dimension;

      const headColumn = {
        header: sanitizeHeaderValue(label || name),
        columns: acc,
        accessorFn: () => "measure",
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
  dimension: string,
  value: string,

  timeConfig: PivotTimeConfig,
) {
  if (isTimeDimension(dimension, timeConfig?.timeDimension)) {
    if (value === "Total") return "Total";
    const timeGrain = getTimeGrainFromDimension(dimension);
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
  totals: PivotDataRow,
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
      name: m,
      format: measure.formatD3 || measure.formatPreset,
    };
  });

  const rowDimensions = rowDimensionNames.map((d) => {
    let label =
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d;
    if (isTimeDimension(d, config.time.timeDimension)) {
      const timeGrain = getTimeGrainFromDimension(d);
      const grainLabel = TIME_GRAIN[timeGrain]?.label || d;
      label = `Time ${grainLabel}`;
    }
    return {
      label,
      name: d,
    };
  });
  const colDimensions = colDimensionNames.map((d) => {
    let label =
      config.allDimensions.find((dimension) => dimension.column === d)?.label ||
      d;
    if (isTimeDimension(d, config.time.timeDimension)) {
      const timeGrain = getTimeGrainFromDimension(d);
      const grainLabel = TIME_GRAIN[timeGrain]?.label || d;
      label = `Time ${grainLabel}`;
    }
    return {
      label,
      name: d,
    };
  });

  let rowDimensionsForColumnDef = rowDimensions;
  let nestedLabel: string;
  if (IsNested) {
    rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
    nestedLabel = rowDimensions.map((d) => d.label || d.name).join(" > ");
  }
  const rowDefinitions: ColumnDef<PivotDataRow>[] =
    rowDimensionsForColumnDef.map((d) => {
      return {
        id: d.name,
        accessorFn: (row) => {
          const value = row[d.name];
          return isNotPivotDataRow(value)
            ? formatRowDimensionValue(d.name, String(value), config.time)
            : value;
        },
        header: nestedLabel,
      };
    });

  const leafColumns: ColumnDef<PivotDataRow>[] = measures.map((m) => {
    return {
      accessorKey: m.name,
      // Using this property to pass the format to the cell
      cell: m.format,
      header: m.label || m.name,
    };
  });

  const groupedColDef = createColumnDefinitionForDimensions(
    config,
    colDimensions,
    columnDimensionAxes || {},
    leafColumns,
    totals,
  );

  return [...rowDefinitions, ...groupedColDef];
}

function isNotPivotDataRow(
  value: unknown,
): value is string | number | undefined {
  return (
    typeof value === "string" ||
    typeof value === "number" ||
    value === undefined
  );
}
