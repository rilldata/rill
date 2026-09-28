import type {
  Expression,
  Schema as MetricsResolverQuery,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * Properties of the `ai` resolver that scope a report's analysis (see `runtime/resolvers/ai.go`).
 * All of them are optional.
 */
export type AIResolverProps = {
  explore?: string;
  dimensions?: string[];
  measures?: string[];
  time_range?: TimeRange;
  comparison_time_range?: TimeRange;
  time_zone?: string;
  where?: Expression;
};

/**
 * Maps the scope of an AI report onto a metrics resolver query, so that {@link mapMetricsResolverQueryToDashboard}
 * can turn it into explore state. The `ai` resolver shares the metrics resolver's time range and filter formats;
 * it only names dimensions and measures as plain strings and points at an explore instead of a metrics view.
 */
export function mapAIResolverPropsToMetricsResolverQuery(
  props: AIResolverProps,
  metricsViewName: string,
): MetricsResolverQuery {
  const query: MetricsResolverQuery = { metrics_view: metricsViewName };
  if (props.dimensions?.length) {
    query.dimensions = props.dimensions.map((name) => ({ name }));
  }
  if (props.measures?.length) {
    query.measures = props.measures.map((name) => ({ name }));
  }
  if (props.time_range) query.time_range = props.time_range;
  if (props.comparison_time_range) {
    query.comparison_time_range = props.comparison_time_range;
  }
  if (props.time_zone) query.time_zone = props.time_zone;
  if (props.where) query.where = props.where;
  return query;
}
