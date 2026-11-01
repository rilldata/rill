import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { ExploreDashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
import { renderInRuntimeContext } from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
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
});
