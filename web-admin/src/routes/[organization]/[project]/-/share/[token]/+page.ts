import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset, token } = await parent();
  const { organization, project } = params;
  const exploreName = token?.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  // Get Explore state from URL params
  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  let urlSearchForPartial = "";
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec) {
    const {
      partialExploreState: partialExploreStateFromUrl,
      urlSearchForPartial: _urlSearchForPartial,
      errors: errorsFromConvert,
    } = convertURLToExploreState(
      exploreName,
      `__${organization}__${project}`,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
    urlSearchForPartial = _urlSearchForPartial;
  }

  return {
    partialExploreState,
    urlSearchForPartial,
    errors,
  };
};
