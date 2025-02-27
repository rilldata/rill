import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, defaultExplorePreset, token } = await parent();
  const exploreName = token?.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  return getExploreStates(
    exploreName,
    `${token.id}__`,
    url.searchParams,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
};
