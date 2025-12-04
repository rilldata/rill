import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";
import { getDashboardFromAggregationRequest } from "@rilldata/web-common/features/explore-mappers/get-dashboard-from-aggregation-request.ts";
import { getDashboardFromComparisonRequest } from "@rilldata/web-common/features/explore-mappers/get-dashboard-from-comparison-request.ts";
import type {
  QueryRequests,
  TransformerArgs,
  TransformerProperties,
} from "@rilldata/web-common/features/explore-mappers/types";
import { convertRequestKeysToCamelCase } from "@rilldata/web-common/features/explore-mappers/utils";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, readable, type Readable } from "svelte/store";

export type MapQueryRequest = {
  exploreName: string;
  queryName?: string;
  queryArgsJson?: string;
  executionTime?: string;
};

export type MapQueryStateOptions = {
  exploreProtoState?: string;
  ignoreFilters?: boolean;
  forceOpenPivot?: boolean;
};

export type MapQueryResponse = {
  isFetching: boolean;
  isLoading: boolean;
  error: Error | null;
  data?: { exploreState: ExploreState; exploreName: string };
};

/**
 * Builds the dashboard url from query name and args.
 * Used to show the relevant dashboard for a report/alert.
 */
export function mapQueryToDashboard(
  { exploreName, queryName, queryArgsJson, executionTime }: MapQueryRequest,
  {
    exploreProtoState,
    ignoreFilters = false,
    forceOpenPivot = false,
  }: MapQueryStateOptions,
): Readable<MapQueryResponse> {
  if (!queryName || !queryArgsJson)
    return readable({
      isFetching: false,
      isLoading: false,
      error: new Error("Required parameters are missing."),
    });

  const queryRequestProperties: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson),
  );

  let metricsViewName: string = "";

  let getDashboardState: (
    args: TransformerArgs<TransformerProperties>,
  ) => Promise<ExploreState>;

  // get metrics view name and the query mapper function based on the query name.
  switch (queryName) {
    case "MetricsViewAggregation":
      metricsViewName =
        (queryRequestProperties as V1MetricsViewAggregationRequest)
          .metricsView ?? "";
      getDashboardState = getDashboardFromAggregationRequest;
      break;

    case "MetricsViewComparison":
      metricsViewName =
        (queryRequestProperties as V1MetricsViewComparisonRequest)
          .metricsViewName ?? "";
      getDashboardState = getDashboardFromComparisonRequest;
      break;

    // TODO
    // case "MetricsViewToplist":
    // case "MetricsViewRows":
    // case "MetricsViewTimeSeries":
  }

  if (!metricsViewName) {
    // error state
    return readable({
      isFetching: false,
      isLoading: false,
      error: new Error(
        "Failed to find metrics view name. Please check the format of the report.",
      ),
    });
  }
  // backwards compatibility for older alerts created on metrics explore directly
  if (!exploreName) exploreName = metricsViewName;

  const instanceId = get(runtime).instanceId;

  return derived(
    [
      useExploreValidSpec(instanceId, exploreName, undefined, queryClient),
      // TODO: handle non-timestamp dashboards
      createQueryServiceMetricsViewTimeRange(
        get(runtime).instanceId,
        metricsViewName,
        {},
        undefined,
        queryClient,
      ),
    ],
    ([validSpecResp, timeRangeSummary], set) => {
      if (validSpecResp.isLoading || timeRangeSummary.isLoading) {
        set({
          isFetching: true,
          isLoading: true,
          error: null,
        });
        return;
      }

      if (validSpecResp.error || timeRangeSummary.error) {
        set({
          isFetching: false,
          isLoading: false,
          error: new Error(
            validSpecResp.error?.response?.data?.message ??
              timeRangeSummary.error?.response?.data?.message,
          ),
        });
        return;
      }

      // Type guard
      if (
        !validSpecResp.data ||
        !validSpecResp.data.explore ||
        !validSpecResp.data.metricsView
      ) {
        set({
          isFetching: false,
          isLoading: false,
          error: new Error("Failed to fetch explore."),
        });
        return;
      }

      // Type guard
      if (!timeRangeSummary.data?.timeRangeSummary) {
        set({
          isFetching: false,
          isLoading: false,
          error: new Error("Failed to fetch time range summary."),
        });
        return;
      }

      const { metricsView, explore } = validSpecResp.data;

      const rillDefaultExploreState = getRillDefaultExploreState(
        validSpecResp.data.metricsView,
        validSpecResp.data.explore,
        timeRangeSummary.data?.timeRangeSummary,
      );
      const exploreStateFromYAMLConfig = getExploreStateFromYAMLConfig(
        validSpecResp.data.explore,
        timeRangeSummary.data?.timeRangeSummary,
      );
      const defaultExploreState = {
        ...rillDefaultExploreState,
        ...exploreStateFromYAMLConfig,
      };
      getDashboardState({
        queryClient,
        instanceId,
        dashboard: defaultExploreState,
        req: queryRequestProperties,
        metricsView,
        explore,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
        exploreProtoState,
        ignoreFilters,
        forceOpenPivot,
      })
        .then((newExploreState) => {
          set({
            isFetching: false,
            isLoading: false,
            error: null,
            data: {
              exploreState: newExploreState,
              exploreName,
            },
          });
        })
        .catch((err) => {
          set({
            isFetching: false,
            isLoading: false,
            error: err.message,
          });
        });
    },
  );
}
