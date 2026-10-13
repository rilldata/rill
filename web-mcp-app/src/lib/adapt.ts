import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import { toVerticalSpec } from "@rilldata/web-common/features/components/charts/cartesian/orientation";
import type { CircularChartSpec } from "@rilldata/web-common/features/components/charts/circular/CircularChartProvider";
import {
  OTHER_VALUE,
  OTHER_VALUE_DOMAIN_KEY,
  TOTAL_DOMAIN_KEY,
} from "@rilldata/web-common/features/components/charts/circular/constants";
import type { ComboChartSpec } from "@rilldata/web-common/features/components/charts/combo/ComboChartProvider";
import type { HeatmapChartSpec } from "@rilldata/web-common/features/components/charts/heatmap/HeatmapChartProvider";
import type { ScatterPlotChartSpec } from "@rilldata/web-common/features/components/charts/scatter/ScatterPlotChartProvider";
import type {
  ChartDataResult,
  ChartDomainValues,
  ChartFieldsMap,
  ChartSpec,
  ChartType,
  FieldConfig,
  TimeDimensionDefinition,
} from "@rilldata/web-common/features/components/charts/types";
import {
  getFieldsByType,
  isFieldConfig,
  isMultiFieldConfig,
} from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { convertISOStringToJSDateWithSameTimeAsSelectedTimeZone } from "@rilldata/web-common/lib/time/timezone";
import {
  V1TimeGrain,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { timeFormat } from "d3-time-format";
import { DateTime, type DurationLikeObject } from "luxon";
import { readable, type Readable } from "svelte/store";
import {
  buildCaptionParts,
  CAPTION_SEPARATOR,
  formatTimeRange,
} from "./caption";
import {
  isCartesianChartType,
  normalizeRows,
  type ChartResult,
  type WireField,
  type WireRow,
  type WireTimeRange,
} from "./contract";
import { topNTarget } from "./controls";

/**
 * Turns a `create_chart` result into what web-common's `Chart.svelte` renders from,
 * plus the title, caption and data table the card shows around it.
 * Pure: no DOM, no network, no host session. Everything the in-app chart gets from
 * TanStack queries and dashboard stores is derived here from the wire payload.
 */

export interface AdaptOptions {
  isDark: boolean;
  /**
   * Zone used when the result does not name one. Rill buckets temporal values in UTC
   * unless the query asked for a zone, so the fallback is UTC rather than the host's zone.
   */
  timeZone?: string;
  locale?: string;
}

export type TableColumnKind = "time" | "dimension" | "measure";

export interface TableColumn {
  key: string;
  label: string;
  kind: TableColumnKind;
  format: (value: unknown) => string;
}

export interface TableModel {
  columns: TableColumn[];
  rows: WireRow[];
  totalRows: number;
}

export interface AdaptedChart {
  chartType: ChartType;
  spec: ChartSpec;
  chartData: Readable<ChartDataResult>;
  /** The same value the store holds; convenient for spec generation and tests. */
  snapshot: ChartDataResult;
  measures: MetricsViewSpecMeasure[];
  /** Project theme variables for the current mode, as `Chart.svelte` expects them. */
  theme: Record<string, string> | undefined;
  title: string;
  /** The formatted time range on its own; the first caption part when the result has one. */
  timeRangeLabel: string | undefined;
  captionParts: string[];
  caption: string;
  table: TableModel;
  rowCount: number;
  timeZone: string;
  timeGrain: V1TimeGrain | undefined;
  temporalField: string | undefined;
  hasComparison: boolean;
  /** Any budget cut the data: a top-N cap or the row limit. */
  truncated: boolean;
  /** The row limit cut the rows; a time series keeps its newest buckets. */
  rowsTruncated: boolean;
  /** The localized banner line for a truncated result. */
  truncationNote: string | undefined;
  /** The server's own explanation of the cut, written for the model. */
  truncationWarning: string | undefined;
  openUrl: string | undefined;
}

export const MAX_TABLE_ROWS = 500;
/** Zero-filling stops when a range would need more buckets than this. */
export const MAX_FILL_BUCKETS = 1500;
export const DEFAULT_TIME_ZONE = "UTC";
const DEFAULT_TIME_GRAIN = V1TimeGrain.TIME_GRAIN_DAY;
/** Label the chart providers use for the time axis; titles key on it ("over Time"). */
const TIME_AXIS_LABEL = "Time";

export function adaptChartResult(
  result: ChartResult,
  options: AdaptOptions,
): AdaptedChart {
  const chartType = result.chart_type;
  const spec = result.spec as unknown as ChartSpec;
  const timeZone =
    result.resolved_time_range?.time_zone ||
    result.spec.time_range?.time_zone ||
    options.timeZone ||
    DEFAULT_TIME_ZONE;
  const locale = options.locale ?? "en";

  const {
    measures: measureNames,
    dimensions: dimensionNames,
    timeDimensions,
  } = getFieldsByType(spec);
  const temporalField = timeDimensions[0];
  const timeGrain = resolveTimeGrain(result, temporalField !== undefined);
  const fields = buildFields(
    result.fields ?? {},
    measureNames,
    dimensionNames,
    timeDimensions,
    timeGrain,
  );
  const measures = measureNames.map(
    (name) => fields[name] as MetricsViewSpecMeasure,
  );

  let rows = normalizeRows(result.data);
  const measureKeys = measureKeysInRows(measureNames, rows);
  rows = rows.map((row) => coerceMeasureValues(row, measureKeys));
  const hasComparison = measureNames.some((name) =>
    measureKeys.includes(name + ComparisonDeltaPreviousSuffix),
  );

  const range = result.resolved_time_range ?? result.spec.time_range;
  if (
    temporalField &&
    timeGrain &&
    range &&
    rows.length > 0 &&
    supportsZeroFill(chartType)
  ) {
    const series = seriesDimension(chartType, spec, temporalField);
    rows = zeroFillTimeBuckets(rows, {
      timeField: temporalField,
      grain: timeGrain,
      range,
      timeZone,
      fillKeys: measureKeys,
      seriesField: series,
      seriesValues: series
        ? stringDomain(result.domain_values?.[series])
        : undefined,
      rowsTruncated: result.rows_truncated === true,
    });
  }
  if (timeDimensions.length > 0) {
    rows = convertTimeFields(rows, timeDimensions, timeZone);
  }

  const domainValues = buildDomainValues(chartType, spec, rows, result, fields);

  const snapshot: ChartDataResult = {
    data: rows as ChartDataResult["data"],
    isFetching: false,
    error: null,
    fields,
    domainValues,
    isDarkMode: options.isDark,
    hasComparison,
    theme: themeManager.resolveThemeColors(result.theme, options.isDark),
  };

  // `create_chart` always sends a title; the humanized type id ("Bar chart") is a last resort.
  const fallbackTitle = chartType.replace(/_/g, " ");
  const title =
    result.title?.trim() ||
    fallbackTitle[0].toUpperCase() + fallbackTitle.slice(1);
  const timeRangeLabel = range
    ? formatTimeRange(range, timeZone, locale, timeGrain)
    : undefined;
  const captionParts = buildCaptionParts({
    range,
    grain: timeGrain,
    showGrain: temporalField !== undefined,
    hasComparison,
    metricsViewLabel:
      result.metrics_view?.display_name || result.metrics_view?.name,
    timeZone,
    locale,
  });

  return {
    chartType,
    spec,
    chartData: readable(snapshot),
    snapshot,
    measures,
    theme: themeManager.resolveThemeObject(result.theme, options.isDark),
    title,
    timeRangeLabel,
    captionParts,
    caption: captionParts.join(CAPTION_SEPARATOR),
    table: buildTable(
      rows,
      fields,
      timeDimensions,
      dimensionNames,
      measureNames,
      hasComparison,
      timeGrain,
    ),
    rowCount: rows.length,
    timeZone,
    timeGrain,
    temporalField,
    hasComparison,
    truncated: result.truncated === true,
    rowsTruncated: result.rows_truncated === true,
    truncationNote: truncationNote(
      result,
      temporalField,
      topNTarget(chartType, result.spec)?.limit,
    ),
    truncationWarning: result.truncation_warning || undefined,
    openUrl: result.open_url || undefined,
  };
}

function resolveTimeGrain(
  result: ChartResult,
  hasTemporalField: boolean,
): V1TimeGrain | undefined {
  const grain = result.time_grain ?? result.spec.time_grain;
  if (grain && grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return grain;
  return hasTemporalField ? DEFAULT_TIME_GRAIN : undefined;
}

function buildFields(
  wireFields: Record<string, WireField>,
  measureNames: string[],
  dimensionNames: string[],
  timeNames: string[],
  grain: V1TimeGrain | undefined,
): ChartFieldsMap {
  const fields: ChartFieldsMap = {};
  for (const name of measureNames) {
    const field = wireFields[name];
    const measure: MetricsViewSpecMeasure = {
      name,
      displayName: field?.display_name || name,
      description: field?.description,
      formatPreset: field?.format_preset,
      formatD3: field?.format_d3,
      formatD3Locale:
        field?.format_d3_locale as MetricsViewSpecMeasure["formatD3Locale"],
      lowerIsBetter: field?.lower_is_better,
    };
    fields[name] = measure;
  }
  for (const name of dimensionNames) {
    const field = wireFields[name];
    const dimension: MetricsViewSpecDimension = {
      name,
      displayName: field?.display_name || name,
      description: field?.description,
    };
    fields[name] = dimension;
  }
  for (const name of timeNames) {
    const definition: TimeDimensionDefinition = grain
      ? {
          field: name,
          displayName: TIME_AXIS_LABEL,
          timeUnit: timeGrainToVegaTimeUnitMap[grain],
          format: TIME_GRAIN[grain]?.d3format,
        }
      : { field: name, displayName: TIME_AXIS_LABEL };
    fields[name] = definition;
  }
  return fields;
}

/** Measure columns present in the rows: the measures plus any `<measure>_prev` companions. */
function measureKeysInRows(measureNames: string[], rows: WireRow[]): string[] {
  const keys = new Set<string>(measureNames);
  const sample = rows.slice(0, 50);
  for (const name of measureNames) {
    const previous = name + ComparisonDeltaPreviousSuffix;
    if (sample.some((row) => previous in row)) keys.add(previous);
  }
  return Array.from(keys);
}

/** Measures arrive as JSON numbers, but large integers may be serialized as strings. */
function coerceMeasureValues(row: WireRow, measureKeys: string[]): WireRow {
  let copy: WireRow | undefined;
  for (const key of measureKeys) {
    const value = row[key];
    if (typeof value === "string" && value.trim() !== "") {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) {
        copy ??= { ...row };
        copy[key] = parsed;
      }
    }
  }
  return copy ?? row;
}

function supportsZeroFill(chartType: ChartType): boolean {
  return isCartesianChartType(chartType) || chartType === "combo_chart";
}

/** The nominal dimension that splits a time series into several series, if any. */
function seriesDimension(
  chartType: ChartType,
  spec: ChartSpec,
  temporalField: string,
): string | undefined {
  const color = (spec as { color?: FieldConfig | string }).color;
  if (isFieldConfig(color) && color.type === "nominal") {
    return color.field;
  }
  if (isCartesianChartType(chartType)) {
    const vertical = toVerticalSpec(spec as CartesianChartSpec);
    const x = vertical.x;
    if (x && x.type === "nominal" && x.field !== temporalField) return x.field;
  }
  return undefined;
}

interface ZeroFillOptions {
  timeField: string;
  grain: V1TimeGrain;
  range: WireTimeRange;
  timeZone: string;
  fillKeys: string[];
  seriesField?: string;
  seriesValues?: string[];
  /**
   * Set when the server cut the rows. It keeps the newest buckets, so the buckets before the
   * first data point were cut rather than empty and must not read as zero.
   */
  rowsTruncated?: boolean;
}

function parseInstant(value: unknown, zone: string): DateTime | undefined {
  let parsed: DateTime | undefined;
  if (typeof value === "string") parsed = DateTime.fromISO(value, { zone });
  else if (typeof value === "number")
    parsed = DateTime.fromMillis(value, { zone });
  else if (value instanceof Date) parsed = DateTime.fromJSDate(value, { zone });
  return parsed?.isValid ? parsed : undefined;
}

/**
 * Adds a zero row for every grain bucket of the time range that has no data, per series.
 * Replaces the aggregation API's `fillMissing`, which the MCP data path does not use.
 * Buckets are walked from the earliest data point so week starts and DST follow the data
 * instead of a locale rule. The fill runs from the range start to the range end; when the
 * server cut the rows it kept the newest buckets, so the fill starts at the first data
 * bucket instead. A top-N cap alone leaves the time axis complete and changes nothing here.
 */
export function zeroFillTimeBuckets(
  rows: WireRow[],
  options: ZeroFillOptions,
): WireRow[] {
  const { timeField, grain, range, timeZone, fillKeys, seriesField } = options;
  const unit = TIME_GRAIN[grain]?.unit;
  if (!unit) return rows;
  const step: DurationLikeObject = { [`${unit}s`]: 1 };

  const rangeStart = DateTime.fromISO(range.start, { zone: timeZone });
  const rangeEnd = DateTime.fromISO(range.end, { zone: timeZone });
  if (!rangeStart.isValid || !rangeEnd.isValid) return rows;

  const present = new Map<string, Set<number>>();
  const seriesSample = new Map<string, unknown>();
  let earliest: DateTime | undefined;
  for (const row of rows) {
    const instant = parseInstant(row[timeField], timeZone);
    if (!instant) continue;
    if (!earliest || instant < earliest) earliest = instant;
    const seriesKey = seriesField ? String(row[seriesField]) : "";
    if (!present.has(seriesKey)) present.set(seriesKey, new Set());
    present.get(seriesKey)!.add(instant.toMillis());
    if (seriesField && !seriesSample.has(seriesKey)) {
      seriesSample.set(seriesKey, row[seriesField]);
    }
  }
  if (!earliest) return rows;

  let cursor = earliest;
  if (!options.rowsTruncated) {
    let steps = 0;
    while (cursor > rangeStart && steps++ < MAX_FILL_BUCKETS) {
      cursor = cursor.minus(step);
    }
  }
  const buckets: DateTime[] = [];
  while (cursor < rangeEnd) {
    buckets.push(cursor);
    if (buckets.length > MAX_FILL_BUCKETS) return rows;
    cursor = cursor.plus(step);
  }

  const series = seriesField
    ? options.seriesValues?.length
      ? options.seriesValues
      : Array.from(present.keys())
    : [""];
  const filled = rows.map((row, index) => ({ row, index }));
  let synthetic = rows.length;
  for (const bucket of buckets) {
    const millis = bucket.toMillis();
    for (const seriesKey of series) {
      if (present.get(seriesKey)?.has(millis)) continue;
      const row: WireRow = { [timeField]: bucket.toUTC().toISO() };
      if (seriesField) {
        row[seriesField] = seriesSample.has(seriesKey)
          ? seriesSample.get(seriesKey)
          : seriesKey;
      }
      for (const key of fillKeys) row[key] = 0;
      filled.push({ row, index: synthetic++ });
    }
  }

  return filled
    .map(({ row, index }) => ({
      row,
      index,
      millis: parseInstant(row[timeField], timeZone)?.toMillis() ?? 0,
    }))
    .sort((a, b) => a.millis - b.millis || a.index - b.index)
    .map(({ row }) => row);
}

/** Mirrors `adjustDataForTimeZone` in web-common's chart data provider. */
function convertTimeFields(
  rows: WireRow[],
  timeFields: string[],
  timeZone: string,
): WireRow[] {
  return rows.map((row) => {
    const copy = { ...row };
    for (const field of timeFields) {
      const value = copy[field];
      const iso =
        typeof value === "string"
          ? value
          : typeof value === "number"
            ? new Date(value).toISOString()
            : value instanceof Date
              ? value.toISOString()
              : undefined;
      if (iso !== undefined) {
        copy[field] = convertISOStringToJSDateWithSameTimeAsSelectedTimeZone(
          iso,
          timeZone,
        );
      }
    }
    return copy;
  });
}

function stringDomain(values: unknown[] | undefined): string[] | undefined {
  if (!Array.isArray(values)) return undefined;
  return values
    .filter((value) => value !== null && value !== undefined)
    .map((value) => String(value));
}

function firstAppearance(rows: WireRow[], field: string): string[] {
  const seen = new Set<string>();
  for (const row of rows) {
    const value = row[field];
    if (value === null || value === undefined) continue;
    seen.add(String(value));
  }
  return Array.from(seen);
}

function measureLabels(names: string[] | undefined, fields: ChartFieldsMap) {
  return (names ?? []).map((name) => fields[name]?.displayName || name);
}

function sumMeasure(rows: WireRow[], field: string): number {
  return rows.reduce((total, row) => {
    const value = row[field];
    return typeof value === "number" && Number.isFinite(value)
      ? total + value
      : total;
  }, 0);
}

/**
 * Ordered domains for the nominal channels: the backend's top-N order when it sent one,
 * else first appearance in the rows. The circular total and the multi-measure legend
 * labels are derived here because the spec generators read them from the same map.
 */
function buildDomainValues(
  chartType: ChartType,
  spec: ChartSpec,
  rows: WireRow[],
  result: ChartResult,
  fields: ChartFieldsMap,
): ChartDomainValues {
  const domain: ChartDomainValues = {};
  for (const [key, values] of Object.entries(result.domain_values ?? {})) {
    if (!Array.isArray(values)) continue;
    const defined = values.filter(
      (value) => value !== null && value !== undefined,
    );
    domain[key] = defined.every((value) => typeof value === "number")
      ? (defined as number[])
      : defined.map((value) => String(value));
  }

  const fallback = (field: FieldConfig | string | undefined) => {
    if (!isFieldConfig(field) || field.type !== "nominal") return;
    if (!(field.field in domain))
      domain[field.field] = firstAppearance(rows, field.field);
  };

  switch (chartType) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized": {
      const vertical = toVerticalSpec(spec as CartesianChartSpec);
      fallback(vertical.x);
      if (isFieldConfig(vertical.color)) {
        if (isMultiFieldConfig(vertical.y)) {
          domain[vertical.color.field] = measureLabels(
            vertical.y.fields,
            fields,
          );
        } else {
          fallback(vertical.color);
        }
      }
      break;
    }
    case "combo_chart": {
      const combo = spec as ComboChartSpec;
      fallback(combo.x);
      if (combo.color?.field && combo.y1?.field && combo.y2?.field) {
        domain[combo.color.field] = measureLabels(
          [combo.y1.field, combo.y2.field],
          fields,
        );
      }
      break;
    }
    case "donut_chart":
    case "pie_chart": {
      const circular = spec as CircularChartSpec;
      fallback(circular.color);
      const measureField = circular.measure?.field;
      if (measureField) {
        const sum = sumMeasure(rows, measureField);
        const total = result.totals?.[measureField] ?? sum;
        domain[TOTAL_DOMAIN_KEY] = [total];
        const other = total - sum;
        if (circular.show_other === true && circular.color && other > 0) {
          domain[OTHER_VALUE_DOMAIN_KEY] = [other];
          const colors = domain[circular.color.field];
          if (Array.isArray(colors) && !colors.includes(OTHER_VALUE as never)) {
            domain[circular.color.field] = [
              ...(colors as string[]),
              OTHER_VALUE,
            ];
          }
        }
      }
      break;
    }
    case "heatmap": {
      const heatmap = spec as HeatmapChartSpec;
      fallback(heatmap.x);
      fallback(heatmap.y);
      break;
    }
    case "scatter_plot": {
      fallback((spec as ScatterPlotChartSpec).color);
      break;
    }
    case "funnel_chart":
      break;
  }
  return domain;
}

/**
 * The banner line for a truncated result. Row truncation keeps the newest buckets of a time
 * series; otherwise the cut is the top-N cap of the chart's nominal channel.
 */
function truncationNote(
  result: ChartResult,
  temporalField: string | undefined,
  topN: number | undefined,
): string | undefined {
  if (result.truncated !== true) return undefined;
  if (result.rows_truncated === true) {
    return temporalField
      ? m.mcp_chart_truncated_recent_periods()
      : m.mcp_chart_truncated();
  }
  return topN === undefined
    ? m.mcp_chart_truncated()
    : m.mcp_chart_truncated_top_values({ n: topN });
}

function buildTable(
  rows: WireRow[],
  fields: ChartFieldsMap,
  timeNames: string[],
  dimensionNames: string[],
  measureNames: string[],
  hasComparison: boolean,
  grain: V1TimeGrain | undefined,
): TableModel {
  const columns: TableColumn[] = [];
  const timeFormatter = timeFormat(
    (grain && TIME_GRAIN[grain]?.d3format) || "%Y-%m-%d %H:%M",
  );
  for (const name of timeNames) {
    columns.push({
      key: name,
      label: fields[name]?.displayName || name,
      kind: "time",
      format: (value) =>
        value instanceof Date
          ? timeFormatter(value)
          : value === null || value === undefined
            ? ""
            : String(value),
    });
  }
  for (const name of dimensionNames) {
    columns.push({
      key: name,
      label: fields[name]?.displayName || name,
      kind: "dimension",
      format: (value) =>
        value === null || value === undefined ? "" : String(value),
    });
  }
  for (const name of measureNames) {
    const measure = fields[name] as MetricsViewSpecMeasure | undefined;
    const formatter = createMeasureValueFormatter<null | undefined>(
      measure ?? { name },
    );
    const format = (value: unknown) => {
      if (typeof value === "number") return formatter(value) ?? "";
      return value === null || value === undefined ? "" : String(value);
    };
    const label = measure?.displayName || name;
    columns.push({ key: name, label, kind: "measure", format });
    if (hasComparison) {
      columns.push({
        key: name + ComparisonDeltaPreviousSuffix,
        label: m.mcp_chart_previous_column({ name: label }),
        kind: "measure",
        format,
      });
    }
  }
  return {
    columns,
    rows: rows.slice(0, MAX_TABLE_ROWS),
    totalRows: rows.length,
  };
}
