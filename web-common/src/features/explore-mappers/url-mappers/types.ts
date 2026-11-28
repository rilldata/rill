import type {
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
  V1MetricsViewRowsRequest,
  V1MetricsViewTimeSeriesRequest,
  V1MetricsViewToplistRequest,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type {
  Schema as MetricsResolverQuery,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

export type QueryRequests =
  | V1MetricsViewAggregationRequest
  | V1MetricsViewToplistRequest
  | V1MetricsViewRowsRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewComparisonRequest;

/**
 * The saved query of an alert or a report.
 */
export type MapQueryRequest = {
  exploreName: string;
  queryName?: string;
  queryArgsJson?: string;
  executionTime?: string;
};

/**
 * How the saved dashboard state of an alert or a report is combined with its query.
 */
export type MapQueryStateOptions = {
  exploreProtoState?: string;
  ignoreFilters?: boolean;
  forceOpenPivot?: boolean;
};

export type URLTransformerArgs<R extends QueryRequests> = {
  client: RuntimeClient;
  req: R;
  // Undefined when the metrics view has no time dimension.
  executionTime?: string;
  urlState?: string;
  ignoreFilters?: boolean;
  forceOpenPivot?: boolean;
};

/**
 * A metrics resolver query, such as one an AI tool call ran,
 * with the time ranges the runtime resolved it to, if known.
 */
export type MetricsResolverQueryMapperArgs = {
  query: MetricsResolverQuery;
  resolvedTimeRange?: TimeRange;
  resolvedComparisonTimeRange?: TimeRange;
};
