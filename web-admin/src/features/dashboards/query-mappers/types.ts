import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type {
  V1ExploreSpec,
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
  V1MetricsViewRowsRequest,
  V1MetricsViewSpec,
  V1MetricsViewTimeSeriesRequest,
  V1MetricsViewToplistRequest,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export type QueryRequests =
  | V1MetricsViewAggregationRequest
  | V1MetricsViewToplistRequest
  | V1MetricsViewRowsRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewComparisonRequest;

export type QueryMapperArgs<R extends QueryRequests> = {
  queryClient: QueryClient;
  instanceId: string;
  dashboard: ExploreState;
  req: R;
  metricsView: V1MetricsViewSpec;
  explore: V1ExploreSpec;
  timeRangeSummary: V1TimeRangeSummary;
  executionTime: string;
  annotations: Record<string, string>;
};
