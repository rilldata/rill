import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
import {
  addZoneOffset,
  removeLocalTimezoneOffset,
} from "@rilldata/web-common/lib/time/timezone";
import type { ColumnDef } from "@tanstack/svelte-table";
import { timeFormat } from "d3-time-format";
import type { ComponentType, SvelteComponent } from "svelte";
import PivotDeltaCell from "./PivotDeltaCell.svelte";
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
  PivotChipType,
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
          const duration = timeGrainToDuration(timeGrain);

          const dt = addZoneOffset(
            removeLocalTimezoneOffset(new Date(value), duration),
            timeConfig?.timeZone,
            duration,
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
      const { name } = dimension;

      const headColumn = {
        id: name,
        header: "",
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
 * Get formatted value for dimension values. Format
 * time dimension values if present.
 */
function formatDimensionValue(
  value: string,
  depth: number,
  timeConfig: PivotTimeConfig,
  rowDimensionNames: string[],
) {
  const dimension = rowDimensionNames?.[depth];
  if (isTimeDimension(dimension, timeConfig?.timeDimension)) {
    if (
      value === "Total" ||
      value === "LOADING_CELL" ||
      value === undefined ||
      value === null
    )
      return value;

    const timeGrain = getTimeGrainFromDimension(dimension);
    const duration = timeGrainToDuration(timeGrain);
    const dt = addZoneOffset(
      removeLocalTimezoneOffset(new Date(value), duration),
      timeConfig?.timeZone,
      duration,
    );
    const timeFormatter = timeFormat(
      timeGrain ? TIME_GRAIN[timeGrain]?.d3format : "%H:%M",
    ) as (d: Date) => string;

    return timeFormatter(dt);
  }
  return value;
}

export type MeasureColumnProps = Array<{
  label: string;
  icon?: ComponentType<SvelteComponent>;
  formatter: (
    value: string | number | null | undefined,
  ) => string | (null | undefined);
  name: string;
  type: MeasureType;
}>;
export function getMeasureColumnProps(
  config: PivotDataStoreConfig,
): MeasureColumnProps {
  const { measureNames } = config;
  return measureNames.map((m) => {
    let measureName = m;
    let label: string = "";
    let icon: ComponentType<SvelteComponent> | undefined;
    let type: MeasureType = "measure";
    if (m.endsWith(COMPARISON_DELTA)) {
      icon = DeltaChange;
      label = "Δ";
      type = "comparison_delta";
      measureName = m.replace(COMPARISON_DELTA, "");
    } else if (m.endsWith(COMPARISON_PERCENT)) {
      icon = DeltaChangePercentage;
      label = "Δ %";
      type = "comparison_percent";
      measureName = m.replace(COMPARISON_PERCENT, "");
    }
    const measure = config.allMeasures.find(
      (measure) => measure.name === measureName,
    );

    if (!measure) {
      console.warn(`Measure ${m} not found in config.allMeasures`);
    }

    return {
      label: label || measure?.displayName || measureName,
      formatter: measure
        ? createMeasureValueFormatter<null | undefined>(measure)
        : (v: string | number | null | undefined) => v?.toString(),
      name: m,
      type,
      icon,
    };
  });
}

export function getDimensionColumnProps(
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
  const { rowDimensionNames, colDimensionNames, isFlat } = config;

  const measures = getMeasureColumnProps(config);
  const rowDimensions = getDimensionColumnProps(rowDimensionNames, config);
  const colDimensions = getDimensionColumnProps(colDimensionNames, config);

  return isFlat
    ? getFlatColumnDef(config, measures, rowDimensions, rowDimensionNames)
    : getNestedColumnDef(
        config,
        measures,
        rowDimensions,
        colDimensions,
        columnDimensionAxes,
        totals,
        rowDimensionNames,
      );
}

function getFlatColumnDef(
  config: PivotDataStoreConfig,
  measures: MeasureColumnProps,
  rowDimensions: Array<{ label: string; name: string }>,
  rowDimensionNames: string[],
): ColumnDef<PivotDataRow>[] {
  const rowDefinitions: ColumnDef<PivotDataRow>[] = rowDimensions.map(
    (d, i) => {
      return {
        id: d.name,
        accessorFn: (row) => row[d.name],
        header: d.label || d.name,
        cell: ({ getValue }) => {
          return formatDimensionValue(
            getValue() as string,
            i,
            config.time,
            rowDimensionNames,
          );
        },
      };
    },
  );

  const leafColumns: ColumnDef<PivotDataRow>[] = measures.map((m) => {
    const formatter = m.formatter;
    return {
      // accessorKey: m.name,
      accessorFn: (row) => formatter(row[m.name]),
      header: m.label || m.name,
      name: m.name,
      meta: {
        icon: m.icon,
      },
      // cell: (info) => {
      //   const measureValue = info.getValue() as number | null | undefined;
      //   if (m.type === "comparison_percent") {
      //     return cellComponent(PercentageChange, {
      //       isNull: measureValue == null,
      //       color: "text-gray-500",
      //       value:
      //         measureValue !== null && measureValue !== undefined
      //           ? formatMeasurePercentageDifference(measureValue)
      //           : null,
      //       inTable: true,
      //     });
      //   } else if (m.type === "comparison_delta") {
      //     return cellComponent(PivotDeltaCell, {
      //       formattedValue: m.formatter(measureValue),
      //       value: measureValue,
      //     });
      //   }
      //   const value = m.formatter(measureValue);

      //   if (value == null) return cellComponent(PivotMeasureCell, {});
      //   return value;
      // },
    };
  });

  const columns = config.pivot.columns;
  const timeDimension = config.time?.timeDimension;

  const measureDefMap = new Map<string, ColumnDef<PivotDataRow>>();
  const dimensionDefMap = new Map<string, ColumnDef<PivotDataRow>>();

  measures.forEach((m, i) => {
    measureDefMap.set(m.name, leafColumns[i]);
  });

  rowDimensions.forEach((d, i) => {
    dimensionDefMap.set(d.name, rowDefinitions[i]);
  });

  // Final column definitions in the order they should appear
  const orderedColumnDefs: ColumnDef<PivotDataRow>[] = [];

  // Process columns in the original order
  columns.forEach((column) => {
    const id = column.id;
    const type = column.type;

    if (type === PivotChipType.Measure) {
      // Add the main measure
      const measureDef = measureDefMap.get(id);
      if (measureDef) {
        orderedColumnDefs.push(measureDef);

        // Add any associated comparison measures right after
        const deltaMeasureName = `${id}${COMPARISON_DELTA}`;
        const deltaMeasureDef = measureDefMap.get(deltaMeasureName);
        if (deltaMeasureDef) {
          orderedColumnDefs.push(deltaMeasureDef);
        }

        const percentMeasureName = `${id}${COMPARISON_PERCENT}`;
        const percentMeasureDef = measureDefMap.get(percentMeasureName);
        if (percentMeasureDef) {
          orderedColumnDefs.push(percentMeasureDef);
        }
      }
    } else {
      let dimensionId = id;
      if (type === PivotChipType.Time) {
        dimensionId = `${timeDimension}_rill_${id}`;
      }

      const dimensionDef = dimensionDefMap.get(dimensionId);
      if (dimensionDef) {
        orderedColumnDefs.push(dimensionDef);
      }
    }
  });

  return orderedColumnDefs;
}

export function getRowNestedLabel(
  rowDimensions: Array<{ label: string; name: string }>,
) {
  return rowDimensions.map((d) => d.label || d.name).join(" > ");
}

function getNestedColumnDef(
  config: PivotDataStoreConfig,
  measures: MeasureColumnProps,
  rowDimensions: Array<{ label: string; name: string }>,
  colDimensions: Array<{ label: string; name: string }>,
  columnDimensionAxes: Record<string, string[]> | undefined,
  totals: PivotDataRow,
  rowDimensionNames: string[],
): ColumnDef<PivotDataRow>[] {
  // For nested tables, we only use the first row dimension in the column definition
  const rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
  const nestedLabel = getRowNestedLabel(rowDimensions);

  // Create row dimension columns
  const rowDefinitions: ColumnDef<PivotDataRow>[] =
    rowDimensionsForColumnDef.map((d) => {
      return {
        id: d.name,
        accessorFn: (row) => row[d.name],
        header: nestedLabel,
        cell: ({ row, getValue }) => {
          const formattedDimensionValue = formatDimensionValue(
            getValue() as string,
            row.depth,
            config.time,
            rowDimensionNames,
          );

          return cellComponent(PivotExpandableCell, {
            value: formattedDimensionValue,
            row,
          });
        },
      };
    });

  let firstDimensionColumns: ColumnDef<PivotDataRow>[] = rowDefinitions;
  if (config.rowDimensionNames.length && config.colDimensionNames.length) {
    firstDimensionColumns = colDimensions.reverse().reduce((acc, dimension) => {
      const { label, name } = dimension;

      const headColumn = {
        id: name,
        header: label || name,
        columns: acc,
      };

      return [headColumn];
    }, rowDefinitions);
  }

  // Create measure columns
  const leafColumns: (ColumnDef<PivotDataRow> & { name: string })[] =
    measures.map((m) => {
      const formatter = m.formatter;
      return {
        accessorKey: m.name,
        // accessorFn: (row) => {
        //   const value = row[m.name];
        //   console.log({ value });
        //   return {
        //     value,
        //     formattedValue: formatter(value),
        //   };
        // },
        header: m.label || m.name,
        name: m.name,
        meta: {
          icon: m.icon,
          type: m.type,
          formatter: m?.formatter,
        },
        // cell: (info) => {
        //   const measureValue = info.getValue() as number | null | undefined;
        //   if (m.type === "comparison_percent") {
        //     return cellComponent(PercentageChange, {
        //       isNull: measureValue == null,
        //       color: "text-gray-500",
        //       value:
        //         measureValue !== null && measureValue !== undefined
        //           ? formatMeasurePercentageDifference(measureValue)
        //           : null,
        //       inTable: true,
        //     });
        //   } else if (m.type === "comparison_delta") {
        //     return cellComponent(PivotDeltaCell, {
        //       formattedValue: m.formatter(measureValue),
        //       value: measureValue,
        //     });
        //   }
        //   const value = m.formatter(measureValue);

        //   if (value == null) return cellComponent(PivotMeasureCell, {});
        //   return value;
        // },
      };
    });

  // Create grouped column definitions
  const groupedColDef = createColumnDefinitionForDimensions(
    config,
    colDimensions,
    columnDimensionAxes || {},
    leafColumns,
    totals,
  );

  return [...firstDimensionColumns, ...groupedColDef];
}
