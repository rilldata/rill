import { getTimeControlsFromURLParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";

export function getLocalComparison(timeFilter: string | undefined) {
  if (!timeFilter) {
    return {
      showLocalTimeComparison: false,
      localComparisonTimeRange: undefined,
    };
  }
  const urlParams = new URLSearchParams(timeFilter);

  const { exploreState, errors } = getTimeControlsFromURLParams(
    urlParams,
    new Map(),
  ); // TODO: this function should not be coupled to ExploreState

  if (errors?.length) {
    return {
      showLocalTimeComparison: false,
      localComparisonTimeRange: undefined,
    };
  }

  return {
    showLocalTimeComparison: !!exploreState.selectedComparisonTimeRange,
    localComparisonTimeRange: exploreState.selectedComparisonTimeRange,
  };
}
