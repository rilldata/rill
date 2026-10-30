import type { FieldConfig } from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaAbsoluteSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type {
  Dimension,
  Expression,
  Measure,
  Schema as MetricsResolverQuery,
  Sort,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * The arguments of a `create_chart` tool call. Rill's own clients send a chart type and Rill's
 * chart spec (see `runtime/metricsview/charts_json_schema.go`); external MCP clients send Flint's
 * `chart_spec` with Rill's data binding in the `data` slot (see `runtime/ai/chartquery`).
 */
export type CreateChartArgs = RillCreateChartArgs | FlintCreateChartArgs;

export type RillCreateChartArgs = {
  chart_type: string;
  spec: CreateChartSpec;
};

/** A Flint encoding: Flint's keys plus Rill's `limit`; only the sort keys matter to an explore. */
export type FlintEncoding = {
  field?: string;
  sortOrder?: "ascending" | "descending";
  sortBy?: string;
  [key: string]: unknown;
};

export type FlintEncodingValue =
  | string
  | FlintEncoding
  | (string | FlintEncoding)[];

export type FlintCreateChartArgs = {
  chart_spec: {
    chartType: string;
    encodings: Record<string, FlintEncodingValue>;
    [key: string]: unknown;
  };
  data: {
    metrics_view: string;
    time_range?: TimeRange;
    comparison_time_range?: TimeRange;
    where?: Expression;
    time_grain?: V1TimeGrain;
  };
};

/** What a field is in the metrics view, as the `fields` of a `create_chart` result report it. */
export type ChartFieldKind = "measure" | "dimension" | "time";

export function isFlintCreateChartArgs(
  args: unknown,
): args is FlintCreateChartArgs {
  if (typeof args !== "object" || args === null) return false;
  const { chart_spec, data } = args as Partial<FlintCreateChartArgs>;
  return (
    typeof chart_spec === "object" &&
    chart_spec !== null &&
    typeof chart_spec.chartType === "string" &&
    typeof data === "object" &&
    data !== null &&
    typeof data.metrics_view === "string"
  );
}

/**
 * A chart spec as the AI builds it: the chart config of `web-common/src/features/components/charts`
 * plus the query scope (`time_range`, `comparison_time_range`, `where`, `time_grain`).
 * The channel keys (`x`, `y`, `y1`, `y2`, `color`, `size`, `dimension`, `stage`, `measure`) hold
 * field configs; the remaining keys are presentation options an explore has no use for.
 */
export type CreateChartSpec = {
  metrics_view: string;
  time_range?: TimeRange;
  comparison_time_range?: TimeRange;
  where?: Expression;
  time_grain?: V1TimeGrain;
  [channel: string]: unknown;
};

/**
 * The spec keys that can hold a field config, in the order their fields are added to the query.
 * `color` may also be a static color string or the `rill_measures` legend of a multi-measure chart,
 * neither of which names a field.
 */
const CHART_CHANNELS = [
  "x",
  "y",
  "y1",
  "y2",
  "color",
  "size",
  "dimension",
  "stage",
  "measure",
] as const;

/**
 * Maps the spec of a `create_chart` call onto a metrics resolver query, so that
 * {@link mapMetricsResolverQueryToDashboard} can turn it into explore state the same way it does
 * for a `query_metrics_view` call. Quantitative fields become measures (all of `fields` for a
 * multi-measure chart), temporal fields become the time dimension floored at the chart's grain,
 * and every other field becomes a dimension. Time ranges and the `where` filter are shared with
 * the resolver's format and pass through unchanged.
 */
export function mapChartSpecToMetricsResolverQuery(
  spec: CreateChartSpec,
): MetricsResolverQuery {
  const grain =
    (spec.time_grain && V1TimeGrainToDateTimeUnit[spec.time_grain]) || "day";
  const dimensions = new Map<string, Dimension>();
  const measures = new Map<string, Measure>();

  for (const channel of CHART_CHANNELS) {
    const field = fieldConfigOf(spec[channel]);
    if (!field) continue;

    if (field.type === "quantitative") {
      for (const name of measureNamesOf(field)) {
        measures.set(name, { name });
      }
    } else if (field.type === "temporal" && field.field) {
      dimensions.set(field.field, {
        name: field.field,
        compute: { time_floor: { dimension: field.field, grain } },
      });
    } else if (field.field) {
      dimensions.set(field.field, { name: field.field });
    }
  }

  // A delta sort orders by a computed measure, which is added to the query here.
  const sort = mapSort(spec, dimensions);
  if (sort?.measure) measures.set(sort.measure.name, sort.measure);

  const query: MetricsResolverQuery = { metrics_view: spec.metrics_view };
  if (dimensions.size) query.dimensions = [...dimensions.values()];
  if (measures.size) query.measures = [...measures.values()];
  if (spec.time_range) query.time_range = spec.time_range;
  if (spec.comparison_time_range) {
    query.comparison_time_range = spec.comparison_time_range;
  }
  if (spec.where) query.where = spec.where;
  if (sort) query.sort = [sort.sort];

  return query;
}

interface MappedSort {
  sort: Sort;
  /** The comparison delta measure a delta sort orders by; the caller adds it to the query. */
  measure?: Measure;
}

/**
 * Maps the first sort found on a channel. Chart sort values name the channel whose field to sort
 * by (`"-y"`, `"x"`, `"-color"`, `"-measure"`), optionally by its comparison delta (`"-y_delta"`);
 * see `ChartSortType`. Array sorts are explicit value orders, which an explore cannot express.
 */
function mapSort(
  spec: CreateChartSpec,
  dimensions: ReadonlyMap<string, Dimension>,
): MappedSort | undefined {
  for (const channel of CHART_CHANNELS) {
    const sort = fieldConfigOf(spec[channel])?.sort;
    if (typeof sort !== "string") continue;

    const desc = sort.startsWith("-");
    const key = desc ? sort.slice(1) : sort;
    const isDelta = key.endsWith(ComparisonDeltaAbsoluteSuffix);
    const target = sortTargetOf(
      spec,
      isDelta ? key.slice(0, -ComparisonDeltaAbsoluteSuffix.length) : key,
    );
    // A multi-measure chart sorts by its first measure.
    const name = target?.fields?.[0] ?? target?.field;
    if (!name) continue;

    // A delta sort needs a comparison period and only works when the explore opens the dimension
    // table, which sorts by the comparison delta natively. With more dimensions the explore opens
    // a pivot, which cannot sort by a computed measure, so sort by the measure itself instead.
    const nominalDimensions = [...dimensions.values()].filter(
      (d) => !d.compute,
    );
    const canSortByDelta =
      isDelta &&
      target?.type === "quantitative" &&
      !!spec.comparison_time_range &&
      nominalDimensions.length === 1;
    if (!canSortByDelta) return { sort: { name, desc } };

    const deltaName = name + ComparisonDeltaAbsoluteSuffix;
    return {
      sort: { name: deltaName, desc },
      measure: {
        name: deltaName,
        compute: { comparison_delta: { measure: name } },
      },
    };
  }
  return undefined;
}

/** Resolves the channel a sort value names to the field config drawn on it. */
function sortTargetOf(
  spec: CreateChartSpec,
  channel: string,
): FieldConfig | undefined {
  switch (channel) {
    case "x":
      return fieldConfigOf(spec.x);
    case "y":
      // Combo charts draw their first measure on y1; funnels sort their stages with "y" too.
      return (
        fieldConfigOf(spec.y) ??
        fieldConfigOf(spec.y1) ??
        fieldConfigOf(spec.measure)
      );
    case "color":
      return fieldConfigOf(spec.color);
    case "measure":
      return fieldConfigOf(spec.measure);
    default:
      return undefined;
  }
}

function fieldConfigOf(value: unknown): FieldConfig | undefined {
  // "value" fields carry a constant such as the `rill_measures` legend, not a metrics view field.
  if (!isFieldConfig(value) || value.type === "value") return undefined;
  return value;
}

function measureNamesOf(field: FieldConfig): string[] {
  return [...(field.fields ?? []), field.field].filter(Boolean);
}

/**
 * Channels in the order their fields are added to the query; channels a chart type adds beyond
 * these follow in the order they were authored.
 */
const FLINT_CHANNELS = [
  "x",
  "y",
  "color",
  "group",
  "column",
  "row",
  "detail",
  "order",
  "size",
  "shape",
  "strokeDash",
  "opacity",
  "x2",
  "y2",
  "open",
  "high",
  "low",
  "close",
  "goal",
  "value",
  "metric",
];

function flintEncodingEntries(
  args: FlintCreateChartArgs,
): [string, FlintEncoding[]][] {
  const encodings = args.chart_spec.encodings ?? {};
  const channels = [
    ...FLINT_CHANNELS.filter((channel) => channel in encodings),
    ...Object.keys(encodings).filter(
      (channel) => !FLINT_CHANNELS.includes(channel),
    ),
  ];
  return channels.map((channel) => {
    const value = encodings[channel];
    const values = Array.isArray(value) ? value : [value];
    return [
      channel,
      values
        .map((entry) => (typeof entry === "string" ? { field: entry } : entry))
        .filter((entry): entry is FlintEncoding => !!entry?.field),
    ];
  });
}

/**
 * Maps the arguments of a Flint-shaped `create_chart` call onto a metrics resolver query. The
 * encodings only name fields, so `kinds` (from the result's `fields`) says which are measures and
 * which is the time dimension, floored at the call's grain; a field the result does not describe
 * is treated as a dimension. The time ranges and the filter come from the data binding.
 */
export function mapFlintChartArgsToMetricsResolverQuery(
  args: FlintCreateChartArgs,
  kinds: Record<string, ChartFieldKind>,
): MetricsResolverQuery {
  const { data } = args;
  const grain =
    (data.time_grain && V1TimeGrainToDateTimeUnit[data.time_grain]) || "day";
  const dimensions = new Map<string, Dimension>();
  const measures = new Map<string, Measure>();
  const entries = flintEncodingEntries(args);

  for (const [, encodings] of entries) {
    for (const { field } of encodings) {
      if (!field) continue;
      switch (kinds[field]) {
        case "measure":
          measures.set(field, { name: field });
          break;
        case "time":
          dimensions.set(field, {
            name: field,
            compute: { time_floor: { dimension: field, grain } },
          });
          break;
        default:
          dimensions.set(field, { name: field });
      }
    }
  }

  const query: MetricsResolverQuery = { metrics_view: data.metrics_view };
  if (dimensions.size) query.dimensions = [...dimensions.values()];
  if (measures.size) query.measures = [...measures.values()];
  if (data.time_range) query.time_range = data.time_range;
  if (data.comparison_time_range) {
    query.comparison_time_range = data.comparison_time_range;
  }
  if (data.where) query.where = data.where;
  const sort = mapFlintSort(entries);
  if (sort) query.sort = [sort];
  return query;
}

/**
 * The first sort an encoding asks for. `sortBy` names a channel (`"y"`) or a field; without it
 * the channel sorts by its own field. A `sortBy` alone means the usual top-values intent, so it
 * is descending unless `sortOrder` says otherwise.
 */
function mapFlintSort(entries: [string, FlintEncoding[]][]): Sort | undefined {
  const fieldOnChannel = (channel: string) =>
    entries.find(([name]) => name === channel)?.[1][0]?.field;
  for (const [channel, encodings] of entries) {
    for (const encoding of encodings) {
      if (!encoding.sortBy && !encoding.sortOrder) continue;
      const name = encoding.sortBy
        ? (fieldOnChannel(encoding.sortBy) ?? encoding.sortBy)
        : fieldOnChannel(channel);
      if (!name) continue;
      const desc = encoding.sortOrder
        ? encoding.sortOrder === "descending"
        : true;
      return { name, desc };
    }
  }
  return undefined;
}
