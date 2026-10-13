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
 * The arguments of a `create_chart` tool call (see `runtime/metricsview/charts_json_schema.go`).
 */
export type CreateChartArgs = {
  chart_type: string;
  spec: CreateChartSpec;
};

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
