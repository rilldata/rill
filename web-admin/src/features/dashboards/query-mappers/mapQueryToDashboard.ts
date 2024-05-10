import { getDashboardFromAggregationRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromAggregationRequest";
import { getDashboardFromComparisonRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromComparisonRequest";
import type {
  QueryMapperArgs,
  QueryRequests,
} from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, readable } from "svelte/store";

type DashboardStateForQuery = {
  state?: string;
  metricsView?: string;
};

/**
 * Builds the dashboard url from query name and args.
 * Used to show the relevant dashboard for a report/alert.
 */
export function mapQueryToDashboard(
  queryName: string | undefined,
  queryArgsJson: string | undefined,
  executionTime: string | undefined,
): CompoundQueryResult<DashboardStateForQuery> {
  if (!queryName)
    return readable({
      isFetching: false,
      error: "",
    });

  let metricsViewName: string = "";
  const req: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson ?? "{}"),
  );
  let getDashboardState: (
    args: QueryMapperArgs<QueryRequests>,
  ) => Promise<MetricsExplorerEntity>;

  // get metrics view name and the query mapper function based on the query name.
  switch (queryName) {
    case "MetricsViewAggregation":
      metricsViewName = (req as V1MetricsViewAggregationRequest).metricsView;
      getDashboardState = getDashboardFromAggregationRequest;
      break;

    case "MetricsViewComparison":
      metricsViewName = (req as V1MetricsViewComparisonRequest).metricsViewName;
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
      error:
        "Failed to find metrics view name. Please check the format of the report.",
    });
  }

  const instanceId = get(runtime).instanceId;

  return derived(
    [
      useMetricsView(instanceId, metricsViewName),
      // TODO: handle non-timestamp dashboards
      createQueryServiceMetricsViewTimeRange(
        get(runtime).instanceId,
        metricsViewName,
        {},
      ),
    ],
    ([metricsViewResource, timeRangeSummary], set) => {
      if (!metricsViewResource.data || !timeRangeSummary.data) {
        set({
          isFetching: true,
          error: "",
        });
        return;
      }

      if (metricsViewResource.error || timeRangeSummary.error) {
        // error state
        set({
          isFetching: false,
          error:
            metricsViewResource.error?.message ??
            timeRangeSummary.error?.message,
        });
        return;
      }

      initLocalUserPreferenceStore(metricsViewName);
      const defaultDashboard = getDefaultMetricsExplorerEntity(
        metricsViewName,
        metricsViewResource.data,
        timeRangeSummary.data,
      );
      getDashboardState({
        queryClient,
        instanceId,
        dashboard: defaultDashboard,
        req,
        metricsView: metricsViewResource.data,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
      })
        .then((newDashboard) => {
          set({
            isFetching: false,
            error: "",
            data: {
              state: getProtoFromDashboardState(newDashboard),
              metricsView: metricsViewName,
            },
          });
        })
        .catch((err) => {
          set({
            isFetching: false,
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
