import { getExploreStores } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { organization, project, dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  return {
    exploreName,
    ...getExploreStores(
      exploreName,
      `${organization}__${project}__`,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ),
  };
};
