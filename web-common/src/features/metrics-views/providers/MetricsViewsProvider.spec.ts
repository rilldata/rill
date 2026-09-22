import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  createTestMetricsViewsProvider,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { describe, expect, it } from "vitest";

const AD_BIDS_MIRROR_METRICS_NAME = "AdBids_mirror_metrics";

useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT,
  // Shares the publisher dimension and the impressions measure with AdBids.
  [AD_BIDS_MIRROR_METRICS_NAME]: {
    ...AD_BIDS_METRICS_INIT,
    dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
    measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
  },
});

describe("MetricsViewsProvider", () => {
  it("loads the spec for each metrics view", async () => {
    const { value: provider, destroy } = await createTestMetricsViewsProvider([
      AD_BIDS_METRICS_NAME,
    ]);

    // Not an exact match: the transport fills in proto default values.
    expect(provider.specs[AD_BIDS_METRICS_NAME]).toMatchObject(
      AD_BIDS_METRICS_INIT,
    );
    expect(provider.measures.map((m) => m.name)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_BID_PRICE_MEASURE,
    ]);
    expect(provider.dimensions.map((d) => d.name)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    // No time dimension, so there is no time range summary to wait for.
    expect(provider.ready).toBe(true);

    destroy();
  });

  it("maps a shared measure or dimension to every metrics view defining it", async () => {
    const { value: provider, destroy } = await createTestMetricsViewsProvider([
      AD_BIDS_METRICS_NAME,
      AD_BIDS_MIRROR_METRICS_NAME,
    ]);

    expect(
      Object.keys(provider.dimensionSpecs[AD_BIDS_PUBLISHER_DIMENSION]),
    ).toEqual([AD_BIDS_METRICS_NAME, AD_BIDS_MIRROR_METRICS_NAME]);
    expect(
      Object.keys(provider.measureSpecs[AD_BIDS_IMPRESSIONS_MEASURE]),
    ).toEqual([AD_BIDS_METRICS_NAME, AD_BIDS_MIRROR_METRICS_NAME]);

    // Only AdBids defines these, so they stay single-entry.
    expect(
      Object.keys(provider.dimensionSpecs[AD_BIDS_DOMAIN_DIMENSION]),
    ).toEqual([AD_BIDS_METRICS_NAME]);
    expect(
      Object.keys(provider.measureSpecs[AD_BIDS_BID_PRICE_MEASURE]),
    ).toEqual([AD_BIDS_METRICS_NAME]);

    // Deduped across metrics views, first definition wins.
    expect(provider.dimensions.map((d) => d.name)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
      AD_BIDS_DOMAIN_DIMENSION,
    ]);

    destroy();
  });
});
