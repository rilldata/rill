import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { organization, project, dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const { partialExploreState, loadedOutsideOfURL, errors } =
    convertURLToExploreState(
      exploreName,
      `__${organization}__${project}`,
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

  console.log(exploreName, urlSearchForPartial);
  return {
    partialExploreState,
    urlSearchForPartial,
    errors,
    exploreName,
  };
};
