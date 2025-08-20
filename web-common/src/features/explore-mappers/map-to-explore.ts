import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getDashboardFromAggregationRequest } from "@rilldata/web-common/features/explore-mappers/getDashboardFromAggregationRequest";
import { getDashboardFromComparisonRequest } from "@rilldata/web-common/features/explore-mappers/getDashboardFromComparisonRequest";
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

type MapQueryResponse = {
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
  exploreName: string,
  queryName: string | undefined,
  queryArgsJson: string | undefined,
  executionTime: string | undefined,
  annotations: Record<string, string>,
  alwaysOpenPivot = false,
): Readable<MapQueryResponse> {
  if (!queryName || !queryArgsJson || !executionTime)
    return readable({
      isFetching: false,
      isLoading: false,
      error: new Error("Required parameters are missing."),
    });

  const queryRequestProperties: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson),
  );

  return mapObjectToExploreState(
    exploreName,
    queryName,
    queryRequestProperties,
    executionTime,
    annotations,
    alwaysOpenPivot,
  );
}

export function mapObjectToExploreState(
  exploreName: string,
  transformerName: string,
  transformerProperties: TransformerProperties,
  executionTime: string,
  annotations: Record<string, string>,
  alwaysOpenPivot = false,
): Readable<MapQueryResponse> {
  if (!executionTime)
    return readable({
      isFetching: false,
      isLoading: false,
      error: new Error("Required parameters are missing."),
    });

  let metricsViewName: string = "";

  let getDashboardState: (
    args: TransformerArgs<TransformerProperties>,
  ) => Promise<ExploreState>;

  // get metrics view name and the query mapper function based on the query name.
  switch (transformerName) {
    case "MetricsViewAggregation":
      metricsViewName =
        (transformerProperties as V1MetricsViewAggregationRequest)
          .metricsView ?? "";
      getDashboardState = getDashboardFromAggregationRequest;
      break;

    case "MetricsViewComparison":
      metricsViewName =
        (transformerProperties as V1MetricsViewComparisonRequest)
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
      useExploreValidSpec(instanceId, exploreName),
      // TODO: handle non-timestamp dashboards
      createQueryServiceMetricsViewTimeRange(
        get(runtime).instanceId,
        metricsViewName,
        {},
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

      const defaultExplorePreset = getDefaultExplorePreset(
        validSpecResp.data.explore,
        validSpecResp.data.metricsView,
        timeRangeSummary.data?.timeRangeSummary,
      );
      const { partialExploreState } = convertPresetToExploreState(
        validSpecResp.data.metricsView,
        validSpecResp.data.explore,
        defaultExplorePreset,
      );
      const defaultExploreState = getFullInitExploreState(
        metricsViewName,
        partialExploreState,
      );
      getDashboardState({
        queryClient,
        instanceId,
        dashboard: defaultExploreState,
        req: transformerProperties,
        metricsView,
        explore,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
        annotations,
        alwaysOpenPivot,
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
