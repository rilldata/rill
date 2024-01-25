import { getSortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1TimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
  type V1MetricsViewRowsRequest,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeSeriesRequest,
  type V1MetricsViewToplistRequest,
  type V1Resource,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, type Readable, readable } from "svelte/store";

type ReportQueryRequest =
  | V1MetricsViewAggregationRequest
  | V1MetricsViewToplistRequest
  | V1MetricsViewRowsRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewComparisonRequest;

type QueryMapperArgs<R extends ReportQueryRequest> = {
  dashboard: MetricsExplorerEntity;
  req: R;
  metricsView: V1MetricsViewSpec;
  timeRangeSummary: V1TimeRangeSummary;
  executionTime: string;
};

type ParsedReportQuery = {
  ready: boolean;
  state?: string;
  metricsView?: string;
  error?: string;
};

/**
 * Reports manually written through file artifacts won't have the UI to feed the url state.
 * Hence we are building the state from the query args in the report.
 */
export function parseReportQuery(
  reportResource: V1Resource,
  executionTime: string,
): Readable<ParsedReportQuery> {
  if (!reportResource?.report?.spec?.queryName)
    return readable({
      ready: false,
    });

  let metricsViewName: string = "";
  const req: ReportQueryRequest = correctRequest(
    JSON.parse(reportResource.report.spec.queryArgsJson),
  );
  let queryMapper: (args: QueryMapperArgs<ReportQueryRequest>) => void;

  // get metrics view name and the query mapper function based on the query name.
  switch (reportResource.report.spec.queryName) {
    case "MetricsViewAggregation":
      metricsViewName = (req as V1MetricsViewAggregationRequest).metricsView;
      queryMapper = mapMetricsViewAggregationRequest;
      break;
    case "MetricsViewToplist":
      metricsViewName = (req as V1MetricsViewToplistRequest).metricsViewName;
      queryMapper = mapMetricsViewToplistRequest;
      break;
    case "MetricsViewRows":
      metricsViewName = (req as V1MetricsViewRowsRequest).metricsViewName;
      queryMapper = mapMetricsViewRowsRequest;
      break;
    case "MetricsViewTimeSeries":
      metricsViewName = (req as V1MetricsViewTimeSeriesRequest).metricsViewName;
      queryMapper = mapMetricsViewTimeSeriesRequest;
      break;
    case "MetricsViewComparison":
      metricsViewName = (req as V1MetricsViewComparisonRequest).metricsViewName;
      queryMapper = mapMetricsViewComparisonRequest;
      break;
  }

  if (!metricsViewName) {
    // error state
    return readable({
      ready: true,
      error:
        "Failed to find metrics view name. Please check the format of the report.",
    });
  }

  return derived(
    [
      useMetaQuery(get(runtime).instanceId, metricsViewName),
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
          ready: false,
        };

      if (metricsViewResource.error || timeRangeSummary.error) {
        // error state
        return {
          ready: true,
          error:
            metricsViewResource.error?.message ??
            timeRangeSummary.error?.message,
        };
      }

      initLocalUserPreferenceStore(metricsViewName);
      const dashboard = getDefaultMetricsExplorerEntity(
        metricsViewName,
        metricsViewResource.data,
        timeRangeSummary.data,
      );
      queryMapper({
        dashboard,
        req,
        metricsView: metricsViewResource.data,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
      });
      console.log(dashboard);
      return {
        ready: true,
        state: getProtoFromDashboardState(dashboard),
        metricsView: metricsViewName,
      };
    },
  );
}

/**
 * This method corrects the underscore naming to camel case.
 * This is the drawback of storing the request object as is.
 */
function correctRequest(req: Record<string, any>): Record<string, any> {
  const newReq: Record<string, any> = {};

  for (const key in req) {
    const newKey = key.replace(/_(\w)/g, (_, c: string) => c.toUpperCase());
    const val = req[key];
    if (val && typeof val === "object" && !("length" in val)) {
      newReq[newKey] = correctRequest(val);
    } else {
      newReq[newKey] = val;
    }
  }

  return newReq;
}

function mapMetricsViewAggregationRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewAggregationRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO
}

function mapMetricsViewToplistRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewToplistRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO
}

function mapMetricsViewRowsRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewRowsRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO
}

function mapMetricsViewTimeSeriesRequest({
  req,
  dashboard,
}: QueryMapperArgs<V1MetricsViewTimeSeriesRequest>) {
  if (req.where) dashboard.whereFilter = req.where;
  // TODO
}

function mapMetricsViewComparisonRequest({
  req,
  dashboard,
  metricsView,
  timeRangeSummary,
  executionTime,
}: QueryMapperArgs<V1MetricsViewComparisonRequest>) {
  if (req.where) dashboard.whereFilter = req.where;

  if (req.timeRange) {
    dashboard.selectedTimeRange = getSelectedTimeRange(
      req.timeRange,
      timeRangeSummary,
      executionTime,
    );
  }

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone;
  }

  if (req.comparisonTimeRange) {
    dashboard.selectedComparisonTimeRange = getSelectedTimeRange(
      req.comparisonTimeRange,
      timeRangeSummary,
      executionTime,
    );

    dashboard.selectedComparisonTimeRange.interval =
      dashboard.selectedTimeRange?.interval;
  }

  dashboard.selectedComparisonDimension = req.dimension.name;
  dashboard.visibleMeasureKeys = new Set(req.measures.map((m) => m.name));

  // if the selected sort is a measure set it to leaderboardMeasureName
  if (
    req.sort?.length &&
    metricsView.measures.findIndex((m) => m.name === req.sort[0].name) >= 0
  ) {
    dashboard.leaderboardMeasureName = req.sort[0].name;
    dashboard.sortDirection = req.sort[0].desc
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
    dashboard.dashboardSortType = getSortType(req.sort[0].sortType);
  }

  dashboard.selectedDimensionName = req.dimension.name;
}

function getSelectedTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  executionTime: string,
) {
  let selectedTimeRange: DashboardTimeControls;

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else {
    selectedTimeRange = isoDurationToFullTimeRange(
      timeRange.isoDuration,
      new Date(timeRangeSummary.min),
      new Date(executionTime),
    );
  }

  selectedTimeRange.interval = timeRange.roundToGrain;

  return selectedTimeRange;
}
