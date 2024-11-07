import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, basePreset, token } = await parent();
  const exploreName = token.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  let partialMetrics: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec) {
    const { entity, errors: errorsFromConvert } = convertURLToMetricsExplore(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      basePreset,
    );
    partialMetrics = entity;
    errors.push(...errorsFromConvert);
  }

  if (
    !(exploreName in get(metricsExplorerStore).entities) &&
    token.state &&
    ![...url.searchParams.keys()].length
  ) {
    // Initial load of the dashboard.
    // Merge home token state to url if present and there are no params in the url
    const newUrl = new URL(url);
    newUrl.searchParams.set("state", token.state);
    throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
  }

  return {
    partialMetrics,
    errors,
  };
};
