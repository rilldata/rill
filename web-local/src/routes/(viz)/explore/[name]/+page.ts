import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, defaultExplorePreset } = await parent();
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec && url) {
    const {
      partialExploreState: partialExploreStateFromUrl,
      errors: errorsFromConvert,
    } = convertURLToExploreState(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
  }

  return {
    partialExploreState,
    errors,
  };
};
