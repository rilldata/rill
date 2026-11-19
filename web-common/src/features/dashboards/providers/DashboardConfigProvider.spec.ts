import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { ExploreDashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
import { renderInRuntimeContext } from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { asyncWait, waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { describe, expect, it } from "vitest";

// Mixed-case resource name, as produced by a file like `AdBids_Mixed_Metrics.yaml`.
const MIXED_CASE_METRICS_NAME = "AdBids_Mixed_Metrics";

const mocks = DashboardFetchMocks.useDashboardFetchMocks();
mocks.mockMetricsView(MIXED_CASE_METRICS_NAME, AD_BIDS_METRICS_INIT);
mocks.mockMetricsExplore(
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  {
    ...AD_BIDS_EXPLORE_INIT,
    // The explore references the metrics view in lowercase. Resource names are case-insensitive
    // in the runtime, so the explore still reconciles and GetExplore returns the mixed-case resource.
    metricsView: MIXED_CASE_METRICS_NAME.toLowerCase(),
  },
  MIXED_CASE_METRICS_NAME,
);

// An explore with a filter config, which the provider hands to the filter bar.
const FILTERED_EXPLORE_NAME = "AdBids_filtered_explore";
const FILTERED_EXPLORE: V1ExploreSpec = {
  ...AD_BIDS_EXPLORE_INIT,
  pinnedFilters: [AD_BIDS_DOMAIN_DIMENSION],
  requiredFilters: [AD_BIDS_PUBLISHER_DIMENSION],
  defaultPreset: {
    where: createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Contoso"]),
  },
};
mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);

describe("ExploreDashboardConfigProvider", () => {
  it("names the metrics view by its resource name rather than the explore's reference", async () => {
    const { value: provider, destroy } = renderInRuntimeContext(
      ({ runtimeClient }) =>
        new ExploreDashboardConfigProvider(runtimeClient, AD_BIDS_EXPLORE_NAME),
    );

    await waitUntil(
      () => provider.metricsViewsProvider.metricsViewNames.length > 0,
      5000,
    );
    expect(provider.metricsViewsProvider.metricsViewNames).toEqual([
      MIXED_CASE_METRICS_NAME,
    ]);

    await waitUntil(() => provider.metricsViewsProvider.ready, 5000);
    expect(
      provider.metricsViewsProvider.specs[MIXED_CASE_METRICS_NAME],
    ).toMatchObject(AD_BIDS_METRICS_INIT);

    destroy();
    provider.cleanup?.();
  });

  it("applies the explore's default, pinned and required filters", async () => {
    mocks.mockMetricsExplore(
      FILTERED_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT,
      FILTERED_EXPLORE,
    );
    const { value: provider, destroy } = renderInRuntimeContext(
      ({ runtimeClient }) =>
        new ExploreDashboardConfigProvider(
          runtimeClient,
          FILTERED_EXPLORE_NAME,
        ),
    );
    const { yamlConfigProvider } = provider;

    await waitUntil(() => yamlConfigProvider.requiredFilters.length > 0, 5000);
    expect(yamlConfigProvider.pinnedFilters).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(yamlConfigProvider.requiredFilters).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(yamlConfigProvider.defaultFilters).toEqual({
      [AD_BIDS_METRICS_NAME]: FILTERED_EXPLORE.defaultPreset?.where,
    });

    destroy();
    provider.cleanup?.();
  });

  it("keeps unsaved pin and required toggles until the filter config itself changes", async () => {
    mocks.mockMetricsExplore(
      FILTERED_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT,
      FILTERED_EXPLORE,
    );
    const {
      value: provider,
      queryClient,
      destroy,
    } = renderInRuntimeContext(
      ({ runtimeClient }) =>
        new ExploreDashboardConfigProvider(
          runtimeClient,
          FILTERED_EXPLORE_NAME,
        ),
    );
    const { yamlConfigProvider } = provider;
    await waitUntil(() => yamlConfigProvider.requiredFilters.length > 0, 5000);

    // Toggled in the editor, not saved to the file yet.
    yamlConfigProvider.togglePinnedFilter(AD_BIDS_PUBLISHER_DIMENSION);
    yamlConfigProvider.toggleRequiredFilter(AD_BIDS_PUBLISHER_DIMENSION);

    // An unrelated edit to the file gives the explore a new spec with the same filter config.
    mocks.mockMetricsExplore(FILTERED_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
      ...FILTERED_EXPLORE,
      displayName: "Renamed",
    });
    await queryClient.refetchQueries();
    // The query notifies its subscribers on a later tick.
    await asyncWait(10);
    // A filter pinned in the editor goes after the ones the file pins.
    expect(yamlConfigProvider.pinnedFilters).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(yamlConfigProvider.requiredFilters).toEqual([]);

    // Once the filter config changes, the file is the source of truth again.
    mocks.mockMetricsExplore(FILTERED_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
      ...FILTERED_EXPLORE,
      pinnedFilters: [],
    });
    await queryClient.refetchQueries();
    await asyncWait(10);
    expect(yamlConfigProvider.pinnedFilters).toEqual([]);
    expect(yamlConfigProvider.requiredFilters).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);

    destroy();
    provider.cleanup?.();
  });
});
