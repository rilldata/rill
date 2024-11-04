import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, basePreset, bookmarks } = await parent();
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  let partialMetrics: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec && url) {
    const { entity, errors: errorsFromConvert } = convertURLToMetricsExplore(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      basePreset,
    );
    partialMetrics = entity;
    errors.push(...errorsFromConvert);
  }

  return {
    defaultPartialMetrics: bookmarks.home?.metricsEntity ?? {},
    partialMetrics,
    errors,
  };
};
