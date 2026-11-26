import { fetchTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges.ts";
import {
  convertAbsoluteTimeRangeToRillTime,
  convertV1TimeRangeToComparisonTimeOption,
  convertV1TimeRangeToTimeFilterParams,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import {
  getRuntimeServiceGetResourceQueryOptions,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

export async function convertAPITimeRangesToUrlParams(
  client: RuntimeClient,
  metricsViewName: string,
  reqTimeRange: V1TimeRange | undefined,
  reqComparisonTimeRange: V1TimeRange | undefined,
  executionTime: string | undefined,
) {
  const urlParams = new URLSearchParams();
  if (!reqTimeRange) return urlParams;

  // Metrics views without a time dimension have no time range to show;
  // the runtime ignores a time range that does not name a time dimension.
  const metricsViewResp = await queryClient.fetchQuery(
    getRuntimeServiceGetResourceQueryOptions(client, {
      name: { kind: ResourceKind.MetricsView, name: metricsViewName },
    }),
  );
  if (!metricsViewResp.resource?.metricsView?.state?.validSpec?.timeDimension) {
    // A range on another time dimension does filter the query,
    // but the explore has no time controls to show it.
    const timeDimension =
      reqTimeRange.timeDimension || reqComparisonTimeRange?.timeDimension;
    if (timeDimension) {
      throw new Error(
        `The explore cannot show this query's time range on "${timeDimension}".`,
      );
    }
    return urlParams;
  }

  const { timeRange, timeZone: reqTimeZone } =
    convertV1TimeRangeToTimeFilterParams(reqTimeRange);
  // The runtime resolves in UTC without a zone.
  // Pin the url to the resolving zone so the dashboard reads the resolved range in the same zone.
  const timeZone = reqTimeZone ?? "UTC";
  if (timeRange) {
    // Resolve time range overriding ref to `executionTime` and set to custom.
    // This keeps the time range consistent regardless of when the link is opened.
    const absTimeRangeResp = await fetchTimeRanges({
      client,
      metricsViewName,
      rillTimes: [timeRange],
      timeZone,
      timeDimension: reqTimeRange.timeDimension,
      executionTime,
    });
    const resolvedAbsTimeRange = absTimeRangeResp.resolvedTimeRanges?.[0];
    if (resolvedAbsTimeRange?.start && resolvedAbsTimeRange.end) {
      urlParams.set(
        ExploreStateURLParams.TimeRange,
        convertAbsoluteTimeRangeToRillTime(
          resolvedAbsTimeRange.start,
          resolvedAbsTimeRange.end,
          timeZone,
        ),
      );
    }
  }

  urlParams.set(ExploreStateURLParams.TimeZone, timeZone);

  if (reqComparisonTimeRange) {
    const comparisonParam = convertV1TimeRangeToComparisonTimeOption(
      reqComparisonTimeRange,
    );
    if (comparisonParam) {
      urlParams.set(ExploreStateURLParams.ComparisonTimeRange, comparisonParam);
    }
  }

  return urlParams;
}
