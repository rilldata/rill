import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
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
  getTimeGrainFromDimension,
  isTimeDimension,
} from "./pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type MeasureType,
  type PivotDataRow,
  type PivotDataStoreConfig,
  type PivotTimeConfig,
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

export function getMeasureColumnProps(config: PivotDataStoreConfig) {
  const { measureNames } = config;
  return measureNames.map((m) => {
    let measureName = m;
    let label: string | undefined;
    let type: MeasureType = "measure";
    if (m.endsWith(COMPARISON_DELTA)) {
      label = "Δ";
      type = "comparison_delta";
      measureName = m.replace(COMPARISON_DELTA, "");
    } else if (m.endsWith(COMPARISON_PERCENT)) {
      label = "Δ %";
      type = "comparison_percent";
      measureName = m.replace(COMPARISON_PERCENT, "");
    }
    const measure = config.allMeasures.find(
      (measure) => measure.name === measureName,
    );

    if (!measure) {
      throw new Error(`Measure ${m} not found in config.allMeasures`);
    }

    return {
      label: label || measure?.displayName || measureName,
      formatter: createMeasureValueFormatter<null | undefined>(measure),
      name: m,
      type,
    };
  });
}

function getDimensionColumnProps(
  dimensionNames: string[],
  config: PivotDataStoreConfig,
) {
  return dimensionNames.map((d) => {
    let label =
      config.allDimensions.find(
        (dimension) => dimension.name === d || dimension.column === d,
      )?.displayName || d;
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

  const { rowDimensionNames, colDimensionNames } = config;

  const measures = getMeasureColumnProps(config);
  const rowDimensions = getDimensionColumnProps(rowDimensionNames, config);
  const colDimensions = getDimensionColumnProps(colDimensionNames, config);

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
        accessorFn: (row) => row[d.name],
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

  const leafColumns: (ColumnDef<PivotDataRow> & { name: string })[] =
    measures.map((m) => {
      return {
        accessorKey: m.name,
        header: m.label || m.name,
        name: m.name,
        cell: (info) => {
          const measureValue = info.getValue() as number | null | undefined;
          if (m.type === "comparison_percent") {
            return cellComponent(PercentageChange, {
              isNull: measureValue == null,
              value:
                measureValue !== null && measureValue !== undefined
                  ? formatMeasurePercentageDifference(measureValue)
                  : null,
              inTable: true,
            });
          }
          const value = m.formatter(measureValue);

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
    totals,
  );

  return [...rowDefinitions, ...groupedColDef];
}
