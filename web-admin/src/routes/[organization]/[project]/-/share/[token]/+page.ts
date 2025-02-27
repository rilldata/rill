import { decompressUrlParams } from "@rilldata/web-common/features/dashboards/url-state/compression";
import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, defaultExplorePreset, token } = await parent();
  const exploreName = token?.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  const searchParams = await decompressUrlParams(url.searchParams);

  return getExploreStates(
    exploreName,
    `${token.id}__`,
    searchParams,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
};
