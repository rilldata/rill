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
import {
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
  type V1MetricsViewRowsRequest,
  type V1MetricsViewTimeSeriesRequest,
  type V1MetricsViewToplistRequest,
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
  ) => MetricsExplorerEntity;

  // get metrics view name and the query mapper function based on the query name.
  switch (queryName) {
    case "MetricsViewAggregation":
      metricsViewName = (req as V1MetricsViewAggregationRequest).metricsView;
      getDashboardState = getDashboardFromAggregationRequest;
      break;
    case "MetricsViewToplist":
      metricsViewName = (req as V1MetricsViewToplistRequest).metricsViewName;
      getDashboardState = getDashboardFromToplistRequest;
      break;
    case "MetricsViewRows":
      metricsViewName = (req as V1MetricsViewRowsRequest).metricsViewName;
      getDashboardState = getDashboardFromRowsRequest;
      break;
    case "MetricsViewTimeSeries":
      metricsViewName = (req as V1MetricsViewTimeSeriesRequest).metricsViewName;
      getDashboardState = getDashboardFromTimeSeriesRequest;
      break;
    case "MetricsViewComparison":
      metricsViewName = (req as V1MetricsViewComparisonRequest).metricsViewName;
      getDashboardState = getDashboardFromComparisonRequest;
      break;
  }

  if (!metricsViewName) {
    // error state
    return readable({
      isFetching: false,
      error:
        "Failed to find metrics view name. Please check the format of the report.",
    });
  }

  return derived(
    [
      useMetricsView(get(runtime).instanceId, metricsViewName),
      // TODO: handle non-timestamp dashboards
      createQueryServiceMetricsViewTimeRange(
        get(runtime).instanceId,
        metricsViewName,
        {},
      ),
    ],
    ([metricsViewResource, timeRangeSummary]) => {
      if (!metricsViewResource.data || !timeRangeSummary.data)
        return {
          isFetching: true,
          error: "",
        };

      if (metricsViewResource.error || timeRangeSummary.error) {
        // error state
        return {
          isFetching: false,
          error:
            metricsViewResource.error?.message ??
            timeRangeSummary.error?.message,
        };
      }

      initLocalUserPreferenceStore(metricsViewName);
      const defaultDashboard = getDefaultMetricsExplorerEntity(
        metricsViewName,
        metricsViewResource.data,
        timeRangeSummary.data,
      );
      const newDashboard = getDashboardState({
        dashboard: defaultDashboard,
        req,
        metricsView: metricsViewResource.data,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
      });
      return {
        isFetching: false,
        error: "",
        data: {
          state: getProtoFromDashboardState(newDashboard),
          metricsView: metricsViewName,
        },
      };
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

function getDashboardFromToplistRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewToplistRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO

  return dashboard;
}

function getDashboardFromRowsRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewRowsRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO

  return dashboard;
}

function getDashboardFromTimeSeriesRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewTimeSeriesRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO

  return dashboard;
}
