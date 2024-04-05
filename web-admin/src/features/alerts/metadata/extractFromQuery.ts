import type {
  V1AlertSpec,
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export function extractFromQuery(alertSpec: V1AlertSpec | undefined): {
  dimension: string | undefined;
  where: V1Expression | undefined;
  having: V1Expression | undefined;
  timeRange: V1TimeRange | undefined;
  comparisonTimeRange: V1TimeRange | undefined;
} {
  if (!alertSpec) {
    return {
      dimension: "",
      where: undefined,
      having: undefined,
      timeRange: undefined,
      comparisonTimeRange: undefined,
    };
  }

  const queryArgs = JSON.parse(alertSpec.queryArgsJson ?? "{}");
  if ("metricsView" in queryArgs) {
    const req = queryArgs as V1MetricsViewAggregationRequest;
    return {
      dimension: req.dimensions?.[0]?.name,
      where: req.where,
      having: req.having,
      timeRange: {
        isoDuration: req.timeRange?.isoDuration,
        isoOffset: req.timeRange?.isoOffset,
      },
      comparisonTimeRange: undefined,
    };
  } else {
    const req = queryArgs as V1MetricsViewComparisonRequest;
    return {
      dimension: req.dimension.name,
      where: req.where,
      having: req.having,
      timeRange: {
        isoDuration: req.timeRange?.isoDuration,
        isoOffset: req.timeRange?.isoOffset,
      },
      comparisonTimeRange: {
        isoDuration: req.comparisonTimeRange?.isoDuration,
        isoOffset: req.comparisonTimeRange?.isoOffset,
      },
    };
  }
}
