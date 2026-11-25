import { fetchTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges.ts";
import {
  convertV1TimeRangeToComparisonTimeOption,
  convertV1TimeRangeToTimeFilterParams,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { type V1TimeRange } from "@rilldata/web-common/runtime-client";
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

  const { timeRange, timeZone } =
    convertV1TimeRangeToTimeFilterParams(reqTimeRange);
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
    if (resolvedAbsTimeRange) {
      urlParams.set(
        ExploreStateURLParams.TimeRange,
        `${resolvedAbsTimeRange.start} to ${resolvedAbsTimeRange.end}`,
      );
    }
  }

  if (timeZone) {
    urlParams.set(ExploreStateURLParams.TimeZone, timeZone);
  }

  if (reqComparisonTimeRange) {
    // The explore url only takes an absolute comparison as `start,end`.
    const comparisonParam =
      reqComparisonTimeRange.start && reqComparisonTimeRange.end
        ? convertAbsoluteTimeRangeToUrlParam(
            reqComparisonTimeRange.start,
            reqComparisonTimeRange.end,
          )
        : convertV1TimeRangeToComparisonTimeOption(
            reqTimeRange,
            reqComparisonTimeRange,
          );
    if (comparisonParam) {
      urlParams.set(ExploreStateURLParams.ComparisonTimeRange, comparisonParam);
    }
  }

  return urlParams;
}

/**
 * Formats an absolute range as the explore url's custom range, `start,end`.
 * Both ends are normalized to ISO strings with milliseconds, which the comparison and highlighted range params require.
 */
export function convertAbsoluteTimeRangeToUrlParam(
  start: string | Date,
  end: string | Date,
) {
  return `${new Date(start).toISOString()},${new Date(end).toISOString()}`;
}
