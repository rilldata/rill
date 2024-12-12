import { getExploreStores } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset, token } = await parent();
  const { organization, project } = params;
  const exploreName = token?.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  return getExploreStores(
    exploreName,
    `${organization}__${project}__`,
    url.searchParams,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
};
