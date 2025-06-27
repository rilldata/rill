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

export interface ComponentTransformerProperties extends ExploreState {
  metricsViewName: string;
}

export type TransformerProperties =
  | QueryRequests
  | ComponentTransformerProperties;

export type TransformerArgs<R extends TransformerProperties> = {
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

export interface ExploreAvailabilityResult {
  isAvailable: boolean;
  exploreName?: string;
  error?: string;
}

export interface DashboardSelectionCriteria {
  preferredType?: "recent" | "most_used" | "first_available";
}

export interface ExploreLinkError {
  type:
    | "VALIDATION_ERROR"
    | "TRANSFORMATION_ERROR"
    | "NETWORK_ERROR"
    | "PERMISSION_ERROR";
  message: string;
  details?: any;
}
