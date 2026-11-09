import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type {
  ChartFieldsMap,
  TimeDimensionDefinition,
} from "@rilldata/web-common/features/components/charts/types";
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
import {
  buildCaptionParts,
  CAPTION_SEPARATOR,
  formatTimeRange,
} from "./caption";
import {
  normalizeRows,
  type ChartResult,
  type WireField,
  type WireFlintChartSpec,
  type WireRow,
  type WireTimeRange,
} from "./contract";
import {
  fieldKind,
  seriesFieldOf,
  temporalFieldOf,
  topNTarget,
} from "./controls";
import {
  buildFlintCompileInput,
  encodingFields,
  type FlintCompileInput,
} from "./flint-input";

/**
 * Turns a `create_chart` result into what the chart surface renders from, Flint's compile input,
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
  /**
   * The chart spec to compile instead of the one the result carries,
   * i.e. the presentation controls' edits (chart type, properties, views) applied client-side.
   */
  chartSpec?: WireFlintChartSpec;
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

/** A result ready to render: Flint's compile input, and what the card and the fullscreen view show around it. */
export interface AdaptedChart {
  chartType: string;
  /** The chart spec in effect: the result's, or the presentation controls' edit of it. */
  chartSpec: WireFlintChartSpec;
  flint: FlintCompileInput;
  /**
   * The rows Flint compiles, with temporal values converted to the resolved time zone and the
   * comparison folded into long form; the table shows the result's own rows.
   */
  rows: WireRow[];
  /** The measures the chart draws, in the shape Rill's value formatters take. */
  measures: MetricsViewSpecMeasure[];
  /** Project theme variables for the current mode, as Rill's Vega theme (`getRillTheme`) takes them. */
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
  /** The chart draws the previous period as a second series; the caption and the comparison control follow it. */
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
/**
 * Flint chart types whose time series are zero-filled, as Rill's own cartesian charts are: a
 * line or an area reads a missing bucket as zero rather than as a straight run across it, and a
 * stack or a group of bars keeps every series in every bucket.
 */
const ZERO_FILL_CHART_TYPES: ReadonlySet<string> = new Set([
  "Line Chart",
  "Area Chart",
  "Streamgraph",
  "Bar Chart",
  "Stacked Bar Chart",
  "Grouped Bar Chart",
]);
export const DEFAULT_TIME_ZONE = "UTC";
const DEFAULT_TIME_GRAIN = V1TimeGrain.TIME_GRAIN_DAY;
/** Label of the time axis, as in Rill's own charts. */
const TIME_AXIS_LABEL = "Time";

/**
 * Adapts a result: the rows are zero-filled and converted to the resolved time zone like
 * Rill's charts do (Flint takes `Date` values), the measures get Rill's formatters, and `flint`
 * holds the compile input for the surface. The title comes from the result, else from the spec's
 * headline; the spec's subtitle leads the caption.
 */
export function adaptResult(
  result: ChartResult,
  options: AdaptOptions,
): AdaptedChart {
  const chartSpec = options.chartSpec ?? result.input.chart_spec;
  const timeZone =
    result.resolved_time_range?.time_zone ||
    result.input.data.time_range?.time_zone ||
    options.timeZone ||
    DEFAULT_TIME_ZONE;
  const locale = options.locale ?? "en";
  const wireFields = result.fields ?? {};
  const timeDimension = result.metrics_view?.time_dimension;

  const measureNames: string[] = [];
  const dimensionNames: string[] = [];
  const timeNames: string[] = [];
  for (const [name, field] of Object.entries(wireFields)) {
    if (field.kind === "measure") measureNames.push(name);
    else if (field.kind === "time" || name === timeDimension)
      timeNames.push(name);
    else dimensionNames.push(name);
  }
  const temporalField = temporalFieldOf(chartSpec, wireFields, timeDimension);
  const timeGrain = resolveTimeGrain(
    result.time_grain ?? result.input.data.time_grain,
    temporalField !== undefined,
  );
  const fields = buildFields(
    wireFields,
    measureNames,
    dimensionNames,
    timeNames,
    timeGrain,
    timeDimension,
  );

  let rows = normalizeRows(result.data);
  const measureKeys = measureKeysInRows(measureNames, rows);
  rows = rows.map((row) => coerceMeasureValues(row, measureKeys));

  const range = result.resolved_time_range ?? result.input.data.time_range;
  const series = seriesFieldOf(chartSpec, wireFields, timeDimension);
  const seriesDomain = series
    ? stringDomain(result.domain_values?.[series])
    : undefined;
  // Filled per series, the comparison's `_prev` values included; rows split by another field
  // (a facet, a detail, a second time field) are left as they are.
  const splits = Object.values(chartSpec.encodings ?? {})
    .flatMap((value) => encodingFields(value))
    .filter(
      (field) =>
        field !== temporalField &&
        fieldKind(field, wireFields, timeDimension) !== "measure",
    );
  if (
    temporalField &&
    timeGrain &&
    range &&
    rows.length > 0 &&
    ZERO_FILL_CHART_TYPES.has(chartSpec.chartType) &&
    splits.every((field) => field === series)
  ) {
    rows = zeroFillTimeBuckets(rows, {
      timeField: temporalField,
      grain: timeGrain,
      range,
      timeZone,
      fillKeys: measureKeys,
      seriesField: series,
      seriesValues: seriesDomain,
      rowsTruncated: result.rows_truncated === true,
    });
  }
  if (timeNames.length > 0) {
    rows = convertTimeFields(rows, timeNames, timeZone);
  }

  // Flint colors the series in the order they first appear in the rows, and rows in time order
  // list the series of a bucket in no particular order: in the server's top-N order, every
  // series keeps its color when a refresh changes the grain or the top-N.
  let flintRows = rows;
  if (temporalField && series && seriesDomain) {
    const rank = new Map(seriesDomain.map((value, index) => [value, index]));
    const position = (row: WireRow) =>
      rank.get(String(row[series])) ?? rank.size;
    flintRows = [...rows].sort((a, b) => position(a) - position(b));
  }
  const flint = buildFlintCompileInput(chartSpec, result, flintRows, timeGrain);
  const otherSlice = combinedSliceLabel(chartSpec, series, seriesDomain);
  // The caption and the comparison control say what the chart draws: the previous period is
  // drawn when it is folded in as a second series. The table shows the previous values it has.
  const hasComparison = flint.comparison !== undefined;
  const hasPreviousValues = measureNames.some((name) =>
    measureKeys.includes(name + ComparisonDeltaPreviousSuffix),
  );

  // Each comparison column is formatted like its measure.
  const measures: MetricsViewSpecMeasure[] = [];
  for (const name of measureNames) {
    const measure = fields[name] as MetricsViewSpecMeasure;
    measures.push(measure);
    const previous = name + ComparisonDeltaPreviousSuffix;
    if (measureKeys.includes(previous)) {
      measures.push({
        ...measure,
        name: previous,
        displayName: flint.fields.field_display_names[previous],
      });
    }
  }

  const title =
    result.title?.trim() || chartSpec.title?.trim() || chartSpec.chartType;
  const timeRangeLabel = range
    ? formatTimeRange(range, timeZone, locale, timeGrain)
    : undefined;
  const subtitle = chartSpec.subtitle?.trim();
  const captionParts = [
    ...(subtitle ? [subtitle] : []),
    ...buildCaptionParts({
      range,
      grain: timeGrain,
      showGrain: temporalField !== undefined,
      hasComparison,
      metricsViewLabel:
        result.metrics_view?.display_name || result.metrics_view?.name,
      timeZone,
      locale,
    }),
  ];

  return {
    chartType: chartSpec.chartType,
    chartSpec,
    flint,
    rows: flint.rows,
    measures,
    theme: themeManager.resolveThemeObject(result.theme, options.isDark),
    title,
    timeRangeLabel,
    captionParts,
    caption: captionParts.join(CAPTION_SEPARATOR),
    table: buildTable(
      rows,
      fields,
      timeNames,
      dimensionNames,
      measureNames,
      hasPreviousValues,
      timeGrain,
    ),
    rowCount: rows.length,
    timeZone,
    timeGrain,
    temporalField,
    hasComparison,
    truncated: result.truncated === true,
    rowsTruncated: result.rows_truncated === true,
    truncationNote: otherSlice
      ? m.mcp_chart_truncated_other({ label: otherSlice })
      : truncationNote(
          result,
          temporalField,
          topNTarget(chartSpec, wireFields, timeDimension)?.limit,
        ),
    truncationWarning: result.truncation_warning || undefined,
    openUrl: result.open_url || undefined,
  };
}

function resolveTimeGrain(
  grain: V1TimeGrain | undefined,
  hasTemporalField: boolean,
): V1TimeGrain | undefined {
  if (grain && grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return grain;
  return hasTemporalField ? DEFAULT_TIME_GRAIN : undefined;
}

/**
 * Rill's field map for the formatters and the table. The metrics view's time dimension is
 * labeled "Time"; another time field keeps its own name.
 */
function buildFields(
  wireFields: Record<string, WireField>,
  measureNames: string[],
  dimensionNames: string[],
  timeNames: string[],
  grain: V1TimeGrain | undefined,
  timeDimension: string | undefined,
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
    const displayName =
      timeDimension === undefined || name === timeDimension
        ? TIME_AXIS_LABEL
        : wireFields[name]?.display_name || name;
    const definition: TimeDimensionDefinition = grain
      ? {
          field: name,
          displayName,
          timeUnit: timeGrainToVegaTimeUnitMap[grain],
          format: TIME_GRAIN[grain]?.d3format,
        }
      : { field: name, displayName };
    fields[name] = definition;
  }
  return fields;
}

/**
 * The label of the slice that combines the values a pie or donut chart's top-N cut: the server
 * appends it to the rows and to the series' domain, past the series' limit, when the measure adds up.
 */
function combinedSliceLabel(
  chartSpec: WireFlintChartSpec,
  series: string | undefined,
  domain: string[] | undefined,
): string | undefined {
  if (!series || !domain) return undefined;
  for (const value of Object.values(chartSpec.encodings ?? {})) {
    if (typeof value !== "object" || Array.isArray(value)) continue;
    if (value.field !== series || value.limit === undefined) continue;
    return domain.length > value.limit ? domain[domain.length - 1] : undefined;
  }
  return undefined;
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
  hasPreviousValues: boolean,
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
    if (hasPreviousValues) {
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
