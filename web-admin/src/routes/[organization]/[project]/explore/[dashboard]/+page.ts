import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, basePreset, bookmarks } = await parent();
  const { dashboard: exploreName } = params;
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
    console.log(partialMetrics.whereFilter);
  }

  if (
    !(exploreName in get(metricsExplorerStore).entities) &&
    bookmarks.home?.metricsEntity &&
    ![...url.searchParams.keys()].length
  ) {
    // Initial load of the dashboard.
    // Merge home bookmark to url if present and there are no params in the url
    const newUrl = new URL(url);
    convertMetricsEntityToURLSearchParams(
      bookmarks.home.metricsEntity,
      newUrl.searchParams,
      exploreSpec,
      basePreset,
    );
    throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
  }

  return {
    partialMetrics,
    errors,
  };
};
