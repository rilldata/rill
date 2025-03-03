import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const exploreName = explore.meta.name.name;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  return {
    resourceName: exploreName,
    ...getExploreStates(
      exploreName,
      undefined,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ),
  };
};
