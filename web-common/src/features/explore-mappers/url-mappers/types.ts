import type {
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
  V1MetricsViewRowsRequest,
  V1MetricsViewTimeSeriesRequest,
  V1MetricsViewToplistRequest,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

type QueryRequests =
  | V1MetricsViewAggregationRequest
  | V1MetricsViewToplistRequest
  | V1MetricsViewRowsRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewComparisonRequest;

export type URLTransformerArgs<R extends QueryRequests> = {
  client: RuntimeClient;
  req: R;
  // Undefined when the metrics view has no time dimension.
  executionTime?: string;
  urlState?: string;
  ignoreFilters?: boolean;
  forceOpenPivot?: boolean;
};
