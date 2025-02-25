import { decompressUrlParams } from "@rilldata/web-common/features/dashboards/url-state/compression";
import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const { organization, project, dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const searchParams = await decompressUrlParams(url.searchParams);

  return {
    exploreName,
    ...getExploreStates(
      exploreName,
      `${organization}__${project}__`,
      searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ),
  };
};
