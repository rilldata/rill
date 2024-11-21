import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, basePreset } = await parent();
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec && url) {
    const {
      partialExploreState: partialExploreStateFromUrl,
      errors: errorsFromConvert,
    } = convertURLToMetricsExplore(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      basePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
  }

  return {
    partialExploreState,
    errors,
  };
};
