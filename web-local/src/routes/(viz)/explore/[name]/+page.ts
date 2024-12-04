import { getPartialExploreStateOrRedirect } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { name: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const { partialExploreState, errors } = getPartialExploreStateOrRedirect(
    exploreName,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
    undefined,
    url,
  );

  return {
    partialExploreState,
    errors,
  };
};
