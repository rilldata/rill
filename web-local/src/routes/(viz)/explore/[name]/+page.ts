import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { name: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const { partialExploreState, loadedOutsideOfURL, errors } =
    convertURLToExploreState(
      exploreName,
      undefined,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
  const urlSearchForPartial = loadedOutsideOfURL
    ? getUpdatedUrlForExploreState(
        exploreSpec,
        defaultExplorePreset,
        partialExploreState,
        url.searchParams,
      )
    : url.searchParams.toString();

  return {
    partialExploreState,
    urlSearchForPartial,
    errors,
    exploreName,
  };
};
