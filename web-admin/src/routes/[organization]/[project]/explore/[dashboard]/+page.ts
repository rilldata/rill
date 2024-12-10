import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { organization, project, dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const { partialExploreState, urlSearchForPartial, errors } =
    convertURLToExploreState(
      exploreName,
      `__${organization}__${project}`,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );

  return {
    partialExploreState,
    urlSearchForPartial,
    errors,
    exploreName,
  };
};
