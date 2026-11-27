import {
  fromTimeRangesParams,
  fromTimeRangeUrlParam,
} from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";

export function getLocalComparison(timeFilter: string | undefined) {
  let showLocalTimeComparison = false;
  let localComparisonTimeRange: DashboardTimeControls | undefined = undefined;
  if (!timeFilter) {
    return { showLocalTimeComparison, localComparisonTimeRange };
  }
  const urlParams = new URLSearchParams(timeFilter);
  const { errors } = fromTimeRangesParams(urlParams, new Map());

  if (errors?.length) {
    return { showLocalTimeComparison, localComparisonTimeRange };
  }

  const compareTimeRange = urlParams.get(
    ExploreStateURLParams.ComparisonTimeRange,
  );
  if (compareTimeRange) {
    localComparisonTimeRange = fromTimeRangeUrlParam(compareTimeRange);
    showLocalTimeComparison = true;
  }

  return { showLocalTimeComparison, localComparisonTimeRange };
}
