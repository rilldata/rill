import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import {
  isHorizontal,
  toVerticalSpec,
} from "@rilldata/web-common/features/components/charts/cartesian/orientation";
import type { CircularChartSpec } from "@rilldata/web-common/features/components/charts/circular/CircularChartProvider";
import type { ComboChartSpec } from "@rilldata/web-common/features/components/charts/combo/ComboChartProvider";
import type { FunnelChartSpec } from "@rilldata/web-common/features/components/charts/funnel/FunnelChartProvider";
import type { HeatmapChartSpec } from "@rilldata/web-common/features/components/charts/heatmap/HeatmapChartProvider";
import type { ScatterPlotChartSpec } from "@rilldata/web-common/features/components/charts/scatter/ScatterPlotChartProvider";
import type {
  ChartDataResult,
  ChartFieldsMap,
  ChartType,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import {
  isFieldConfig,
  isMultiFieldConfig,
} from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  WorkbookChartBlock,
  WorkbookChartSeries,
  WorkbookChartType,
  WorkbookColumnType,
  WorkbookSeriesMark,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type {
  MetricsViewSpecMeasure,
  V1MetricsViewAggregationResponseDataItem,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { excelNumberFormat } from "./number-format";
import type { ExportColumn, ExportTable } from "./payload";
import { excelDateFormat, localIsoInZone, timeColumnType } from "./time-format";

export interface ChartExportInput {
  chartType: ChartType;
  spec: CanvasChartSpec;
  data: ChartDataResult;
  timeGrain: V1TimeGrain | undefined;
  timeZone: string;
}

export interface ChartExport {
  table: ExportTable;
  // The native chart bound to the table, without its data sheet id (assigned
  // by the request builder). Undefined when the chart type has no Excel
  // equivalent, in which case the block is a snapshot of the component.
  chart: WorkbookChartBlock | undefined;
}

const NATIVE_CHART_TYPES: ReadonlySet<ChartType> = new Set<ChartType>([
  "bar_chart",
  "line_chart",
  "area_chart",
  "stacked_bar",
  "stacked_bar_normalized",
  "pie_chart",
  "donut_chart",
  "scatter_plot",
  "combo_chart",
]);

// Legends with more entries than this go on the right, where a tall list fits
// better than a wrapped band at the bottom.
const MAX_HORIZONTAL_LEGEND_ENTRIES = 6;

// Whether a canvas chart type maps to a native Excel chart. Heatmaps and
// funnels don't. A combo chart always does: its colour field only names the
// legend entries of its two measures and never splits the data by a dimension
// (see ComboChartProvider).
export function hasNativeChart(type: ChartType): boolean {
  return NATIVE_CHART_TYPES.has(type);
}

// Maps a rendered chart to the data sheet behind it and, where Excel has an
// equivalent, a native chart bound to that sheet.
export function mapChart(input: ChartExportInput): ChartExport {
  const mapped = mapByType(input);
  // A chart with nothing to plot (no rows, or a split with no values and so no
  // series) has no native chart: its block says "No data", like the canvas does.
  if (
    mapped.chart &&
    (mapped.table.rows.length === 0 || mapped.chart.series.length === 0)
  ) {
    return { table: mapped.table, chart: undefined };
  }
  return mapped;
}

function mapByType(input: ChartExportInput): ChartExport {
  switch (input.chartType) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized":
      return mapCartesian(input);
    case "pie_chart":
    case "donut_chart":
      return mapCircular(input);
    case "scatter_plot":
      return mapScatter(input);
    case "combo_chart":
      return mapCombo(input);
    case "heatmap":
      return mapHeatmap(input);
    case "funnel_chart":
      return mapFunnel(input);
  }
}

function mapCartesian(input: ChartExportInput): ChartExport {
  const { chartType, data, timeGrain, timeZone } = input;
  const raw = input.spec as CartesianChartSpec;
  // Normalize to dimension on x and measure on y; see orientation.ts.
  const spec = toVerticalSpec(raw);
  const x = spec.x;
  const y = spec.y;
  const measures = measureFields(y);
  const colorField =
    isFieldConfig(spec.color) && spec.color.type === "nominal"
      ? spec.color.field
      : undefined;

  if (!x?.field || measures.length === 0) {
    return {
      table: longTable([], measures, data, timeGrain, timeZone),
      chart: undefined,
    };
  }

  const categoryColumn = fieldColumn(x, data.fields, timeGrain);
  const table = sortByTime(
    colorField && measures.length === 1
      ? wideTable(categoryColumn, x, colorField, measures[0], data, timeZone)
      : longTable([x], measures, data, timeGrain, timeZone),
  );
  const series = table.columns
    .slice(1)
    .map((column) => new WorkbookChartSeries({ column: column.name }));

  const horizontal = isHorizontal(raw);
  const barType = horizontal ? WorkbookChartType.BAR : WorkbookChartType.COLUMN;
  let type = barType;
  let stacked = false;
  let percentStacked = false;
  switch (chartType) {
    case "stacked_bar":
      stacked = true;
      break;
    case "stacked_bar_normalized":
      percentStacked = true;
      break;
    case "line_chart":
      type = WorkbookChartType.LINE;
      break;
    case "area_chart":
      type = WorkbookChartType.AREA;
      stacked = !!colorField;
      break;
  }

  const hasLegend = !!colorField || series.length > 1;
  const measure = measureSpec(data.fields, measures[0]);
  return {
    table,
    chart: new WorkbookChartBlock({
      type,
      categoryColumn: categoryColumn.name,
      series,
      legendPosition: hasLegend
        ? fitLegend(
            legendPosition(isFieldConfig(spec.color) ? spec.color : undefined),
            series.length,
          )
        : "none",
      xAxisTitle: x.showAxisTitle ? label(data.fields, x.field) : "",
      yAxisTitle:
        y?.showAxisTitle && measures.length === 1
          ? label(data.fields, measures[0])
          : "",
      stacked,
      percentStacked,
      numberFormat: excelNumberFormat(
        measure,
        columnValues(table, measures[0]),
      ),
    }),
  };
}

function mapCircular(input: ChartExportInput): ChartExport {
  const { chartType, data, timeGrain, timeZone } = input;
  const spec = input.spec as CircularChartSpec;
  const dimension = spec.color;
  const measure = spec.measure?.field;
  const table = longTable(
    dimension?.field ? [dimension] : [],
    measure ? [measure] : [],
    data,
    timeGrain,
    timeZone,
  );
  if (!dimension?.field || !measure) return { table, chart: undefined };
  return {
    table,
    chart: new WorkbookChartBlock({
      type:
        chartType === "donut_chart"
          ? WorkbookChartType.DOUGHNUT
          : WorkbookChartType.PIE,
      categoryColumn: dimension.field,
      series: [new WorkbookChartSeries({ column: measure })],
      legendPosition: fitLegend(
        legendPosition(dimension, "right"),
        table.rows.length,
      ),
      numberFormat: excelNumberFormat(
        measureSpec(data.fields, measure),
        columnValues(table, measure),
      ),
    }),
  };
}

function mapScatter(input: ChartExportInput): ChartExport {
  const { data, timeGrain, timeZone } = input;
  const spec = input.spec as ScatterPlotChartSpec;
  const dimensions: FieldConfig[] = [];
  if (spec.dimension?.field) dimensions.push(spec.dimension);
  if (isFieldConfig(spec.color) && spec.color.field)
    dimensions.push(spec.color);
  const axes = [spec.x, spec.y].filter((f): f is FieldConfig => !!f?.field);
  const timeAxes = axes.filter((f) => f.type === "temporal");
  const measures = axes
    .filter((f) => f.type === "quantitative")
    .map((f) => f.field);
  if (spec.size?.field) measures.push(spec.size.field);

  const table = longTable(
    [...dimensions, ...timeAxes],
    measures,
    data,
    timeGrain,
    timeZone,
  );
  if (!spec.x?.field || !spec.y?.field) return { table, chart: undefined };
  return {
    table,
    chart: new WorkbookChartBlock({
      type: WorkbookChartType.SCATTER,
      categoryColumn: spec.x.field,
      series: [new WorkbookChartSeries({ column: spec.y.field })],
      legendPosition: "none",
      xAxisTitle: spec.x.showAxisTitle ? label(data.fields, spec.x.field) : "",
      yAxisTitle: spec.y.showAxisTitle ? label(data.fields, spec.y.field) : "",
      numberFormat: excelNumberFormat(
        measureSpec(data.fields, spec.y.field),
        columnValues(table, spec.y.field),
      ),
    }),
  };
}

function mapCombo(input: ChartExportInput): ChartExport {
  const { data, timeGrain, timeZone } = input;
  const spec = input.spec as ComboChartSpec;
  const measures = [spec.y1?.field, spec.y2?.field].filter(
    (f): f is string => !!f,
  );
  const table = sortByTime(
    longTable(
      spec.x?.field ? [spec.x] : [],
      measures,
      data,
      timeGrain,
      timeZone,
    ),
  );
  if (!spec.x?.field || !spec.y1?.field) {
    return { table, chart: undefined };
  }

  const series = [
    new WorkbookChartSeries({
      column: spec.y1.field,
      mark:
        spec.y1.mark === "line"
          ? WorkbookSeriesMark.LINE
          : WorkbookSeriesMark.BAR,
    }),
  ];
  if (spec.y2?.field) {
    series.push(
      new WorkbookChartSeries({
        column: spec.y2.field,
        mark:
          spec.y2.mark === "bar"
            ? WorkbookSeriesMark.BAR
            : WorkbookSeriesMark.LINE,
        secondaryAxis: true,
      }),
    );
  }
  return {
    table,
    chart: new WorkbookChartBlock({
      type: WorkbookChartType.COMBO,
      categoryColumn: spec.x.field,
      series,
      legendPosition: legendPosition(
        isFieldConfig(spec.color) ? spec.color : undefined,
      ),
      xAxisTitle: spec.x.showAxisTitle ? label(data.fields, spec.x.field) : "",
      yAxisTitle: spec.y1.showAxisTitle
        ? label(data.fields, spec.y1.field)
        : "",
      numberFormat: excelNumberFormat(
        measureSpec(data.fields, spec.y1.field),
        columnValues(table, spec.y1.field),
      ),
    }),
  };
}

function mapHeatmap(input: ChartExportInput): ChartExport {
  const { data, timeGrain, timeZone } = input;
  const spec = input.spec as HeatmapChartSpec;
  const dimensions = [spec.x, spec.y].filter(
    (f): f is FieldConfig => !!f?.field,
  );
  return {
    table: sortByTime(
      longTable(
        dimensions,
        spec.color?.field ? [spec.color.field] : [],
        data,
        timeGrain,
        timeZone,
      ),
    ),
    chart: undefined,
  };
}

function mapFunnel(input: ChartExportInput): ChartExport {
  const { data, timeGrain, timeZone } = input;
  const spec = input.spec as FunnelChartSpec;
  const measures = measureFields(spec.measure);
  const stage =
    spec.breakdownMode !== "measures" && spec.stage?.field ? [spec.stage] : [];
  return {
    table: longTable(stage, measures, data, timeGrain, timeZone),
    chart: undefined,
  };
}

// A table with one row per data row: the dimension and time columns, then the
// measures, then the previous-period value of each measure the data carries.
function longTable(
  dimensions: FieldConfig[],
  measures: string[],
  data: ChartDataResult,
  timeGrain: V1TimeGrain | undefined,
  timeZone: string,
): ExportTable {
  const columns: ExportColumn[] = dimensions.map((field) =>
    fieldColumn(field, data.fields, timeGrain),
  );
  for (const measure of measures) {
    columns.push(measureColumn(measure, data.fields, data.data));
  }
  if (data.hasComparison) {
    for (const measure of measures) {
      const previous = measure + ComparisonDeltaPreviousSuffix;
      if (data.data.some((row) => previous in row)) {
        columns.push(
          measureColumn(previous, data.fields, data.data, {
            displayName: `${label(data.fields, measure)} (previous)`,
            spec: measureSpec(data.fields, measure),
          }),
        );
      }
    }
  }
  const rows = data.data.map((row) =>
    columns.map((column) => cellValue(row[column.name], column, timeZone)),
  );
  return { columns, rows };
}

// A table pivoted wide for a chart split by a dimension: one row per category
// and one column per split value, so a single Excel chart can reference it.
function wideTable(
  categoryColumn: ExportColumn,
  category: FieldConfig,
  colorField: string,
  measure: string,
  data: ChartDataResult,
  timeZone: string,
): ExportTable {
  const domain = data.domainValues?.[colorField];
  const colorValues: string[] = domain
    ? domain.map(String)
    : uniqueInOrder(data.data.map((row) => String(row[colorField])));
  const measureSpecValue = measureSpec(data.fields, measure);

  const byCategory = new Map<
    string,
    { category: unknown; values: Map<string, unknown> }
  >();
  for (const row of data.data) {
    const categoryValue = cellValue(
      row[category.field],
      categoryColumn,
      timeZone,
    );
    const key = String(categoryValue);
    let entry = byCategory.get(key);
    if (!entry) {
      entry = { category: categoryValue, values: new Map() };
      byCategory.set(key, entry);
    }
    entry.values.set(String(row[colorField]), row[measure]);
  }

  const columns: ExportColumn[] = [
    categoryColumn,
    ...colorValues.map((value, i) => ({
      name: `${measure}_${i}`,
      displayName: value,
      type: WorkbookColumnType.NUMBER,
      numberFormat: excelNumberFormat(
        measureSpecValue,
        data.data.map((row) => row[measure]),
      ),
    })),
  ];
  const rows = Array.from(byCategory.values()).map((entry) => [
    entry.category,
    ...colorValues.map((value) => entry.values.get(value) ?? null),
  ]);
  return { columns, rows };
}

// Puts the rows of a table whose first column holds time in time order, oldest
// first, as a time axis draws them: chart data does not always arrive in that
// order. Rows without a time keep their place at the end.
function sortByTime(table: ExportTable): ExportTable {
  const type = table.columns[0]?.type;
  if (
    type !== WorkbookColumnType.DATE &&
    type !== WorkbookColumnType.TIMESTAMP
  ) {
    return table;
  }
  const time = (row: unknown[]) =>
    typeof row[0] === "string" ? row[0] : undefined;
  const rows = [...table.rows].sort((a, b) => {
    const [ta, tb] = [time(a), time(b)];
    if (ta === undefined || tb === undefined) {
      return (ta === undefined ? 1 : 0) - (tb === undefined ? 1 : 0);
    }
    return ta < tb ? -1 : ta > tb ? 1 : 0;
  });
  return { ...table, rows };
}

function measureFields(field: FieldConfig | undefined): string[] {
  if (!field) return [];
  const names = isMultiFieldConfig(field) ? [...(field.fields ?? [])] : [];
  if (
    field.type === "quantitative" &&
    field.field &&
    !names.includes(field.field)
  ) {
    names.unshift(field.field);
  }
  return names;
}

function fieldColumn(
  field: FieldConfig,
  fields: ChartFieldsMap,
  timeGrain: V1TimeGrain | undefined,
): ExportColumn {
  if (field.type === "temporal") {
    return {
      name: field.field,
      displayName: label(fields, field.field),
      type: timeColumnType(timeGrain),
      numberFormat: excelDateFormat(timeGrain),
    };
  }
  return {
    name: field.field,
    displayName: label(fields, field.field),
    type: WorkbookColumnType.STRING,
  };
}

function measureColumn(
  name: string,
  fields: ChartFieldsMap,
  rows: V1MetricsViewAggregationResponseDataItem[],
  override?: { displayName: string; spec: MetricsViewSpecMeasure | undefined },
): ExportColumn {
  return {
    name,
    displayName: override?.displayName ?? label(fields, name),
    type: WorkbookColumnType.NUMBER,
    numberFormat: excelNumberFormat(
      override?.spec ?? measureSpec(fields, name),
      rows.map((row) => row[name]),
    ),
  };
}

function cellValue(
  value: unknown,
  column: ExportColumn,
  timeZone: string,
): unknown {
  if (
    column.type === WorkbookColumnType.DATE ||
    column.type === WorkbookColumnType.TIMESTAMP
  ) {
    return localIsoInZone(value, timeZone);
  }
  return value ?? null;
}

function columnValues(table: ExportTable, name: string): unknown[] {
  const index = table.columns.findIndex((column) => column.name === name);
  return index === -1 ? [] : table.rows.map((row) => row[index]);
}

function label(fields: ChartFieldsMap, name: string): string {
  const field = fields[name];
  if (field && "displayName" in field && field.displayName) {
    return field.displayName;
  }
  return name;
}

function measureSpec(
  fields: ChartFieldsMap,
  name: string,
): MetricsViewSpecMeasure | undefined {
  const field = fields[name];
  // Dimensions and time definitions never carry a format; measures may.
  if (
    field &&
    ("formatPreset" in field || "formatD3" in field || "expression" in field)
  ) {
    return field as MetricsViewSpecMeasure;
  }
  return undefined;
}

function legendPosition(
  color: FieldConfig | undefined,
  fallback = "bottom",
): string {
  const orientation =
    color && "legendOrientation" in color ? color.legendOrientation : undefined;
  return orientation ?? fallback;
}

// Moves a top or bottom legend to the right when it has too many entries to
// fit in one band.
function fitLegend(position: string, entries: number): string {
  if (
    (position === "top" || position === "bottom") &&
    entries > MAX_HORIZONTAL_LEGEND_ENTRIES
  ) {
    return "right";
  }
  return position;
}

function uniqueInOrder(values: string[]): string[] {
  return Array.from(new Set(values));
}
