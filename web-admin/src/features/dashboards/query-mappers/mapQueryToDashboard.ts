import { getDashboardFromAggregationRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromAggregationRequest";
import { getDashboardFromComparisonRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromComparisonRequest";
import type {
  QueryMapperArgs,
  QueryRequests,
} from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, readable, type Readable } from "svelte/store";

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
): Readable<{
  isFetching: boolean;
  isLoading: boolean;
  error: Error;
  data?: { exploreState: MetricsExplorerEntity; exploreName: string };
}> {
  if (!queryName || !queryArgsJson || !executionTime)
    return readable({
      isFetching: false,
      isLoading: false,
      error: new Error("Required parameters are missing."),
    });

  let metricsViewName: string = "";
  const req: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson),
  );
  let getDashboardState: (
    args: QueryMapperArgs<QueryRequests>,
  ) => Promise<MetricsExplorerEntity>;

  // get metrics view name and the query mapper function based on the query name.
  switch (queryName) {
    case "MetricsViewAggregation":
      metricsViewName =
        (req as V1MetricsViewAggregationRequest).metricsView ?? "";
      getDashboardState = getDashboardFromAggregationRequest;
      break;

    case "MetricsViewComparison":
      metricsViewName =
        (req as V1MetricsViewComparisonRequest).metricsViewName ?? "";
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
          error: new Error(""),
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
      if (!timeRangeSummary.data) {
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
        req,
        metricsView,
        explore,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
        annotations,
      })
        .then((newExploreState) => {
          set({
            isFetching: false,
            isLoading: false,
            error: new Error(),
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

/**
 * This method corrects the underscore naming to camel case.
 * This is the drawback of storing the request object as is.
 */
function convertRequestKeysToCamelCase(
  req: Record<string, any>,
): Record<string, any> {
  const newReq: Record<string, any> = {};

  for (const key in req) {
    const newKey = key.replace(/_(\w)/g, (_, c: string) => c.toUpperCase());
    const val = req[key];
    if (val && typeof val === "object" && !("length" in val)) {
      newReq[newKey] = convertRequestKeysToCamelCase(val);
    } else {
      newReq[newKey] = val;
    }
  }

  return newReq;
}
