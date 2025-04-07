import { getTimeControlsFromURLParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";

export function getLocalComparison(timeFilter: string | undefined) {
  let showLocalTimeComparison = false;
  let localComparisonTimeRange: DashboardTimeControls | undefined = undefined;
  if (!timeFilter) {
    return { showLocalTimeComparison, localComparisonTimeRange };
  }
  const urlParams = new URLSearchParams(timeFilter);

  const { exploreState } = getTimeControlsFromURLParams(urlParams, new Map()); // TODO: this function should not be coupled to ExploreState

  return {
    showLocalTimeComparison: !!exploreState.selectedComparisonTimeRange,
    localComparisonTimeRange: exploreState.selectedComparisonTimeRange,
  };
}
