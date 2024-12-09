import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { name: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const { partialExploreState, loaded, errors } = convertURLToExploreState(
    exploreName,
    undefined,
    url.searchParams,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );

  return {
    partialExploreState,
    loaded,
    errors,
  };
};
