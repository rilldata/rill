import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  AD_BIDS_BASE_FILTER,
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_CLEARED_FILTER,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXCLUDE_FILTER,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_WITH_DELETED_DIMENSION,
  ALL_TIME_PARSED_TEST_CONTROLS,
  CUSTOM_TEST_CONTROLS,
  LAST_6_HOURS_TEST_CONTROLS,
  LAST_6_HOURS_TEST_PARSED_CONTROLS,
  assertMetricsView,
  clearMetricsExplorerStore,
  createAdBidsInStore,
  createAdBidsMirrorInStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores-test-data";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";

describe("dashboard-stores", () => {
  beforeEach(() => {
    clearMetricsExplorerStore();
  });

  it("Toggle filters", () => {
    createAdBidsInStore();
    assertMetricsView(AD_BIDS_NAME);

    // add filters
    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
      "Google"
    );
    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook"
    );
    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_DOMAIN_DIMENSION,
      "google.com"
    );
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_BASE_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_BASE_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS
    );

    // toggle to exclude
    metricsExplorerStore.toggleFilterMode(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION
    );
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_EXCLUDE_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_EXCLUDE_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS
    );

    // clear for Domain
    metricsExplorerStore.clearFilterForDimension(
      AD_BIDS_NAME,
      AD_BIDS_DOMAIN_DIMENSION,
      true
    );
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_CLEARED_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_CLEARED_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS
    );

    // clear
    metricsExplorerStore.clearFilters(AD_BIDS_NAME);
    assertMetricsView(AD_BIDS_NAME);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      ALL_TIME_PARSED_TEST_CONTROLS
    );
  });

  it("Update time selections", () => {
    createAdBidsInStore();
    assertMetricsView(AD_BIDS_NAME);

    // select a different time
    metricsExplorerStore.setSelectedTimeRange(
      AD_BIDS_NAME,
      LAST_6_HOURS_TEST_CONTROLS
    );
    assertMetricsView(AD_BIDS_NAME, undefined, LAST_6_HOURS_TEST_CONTROLS);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are not persisted
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      LAST_6_HOURS_TEST_PARSED_CONTROLS
    );

    // select custom time
    metricsExplorerStore.setSelectedTimeRange(
      AD_BIDS_NAME,
      CUSTOM_TEST_CONTROLS
    );
    assertMetricsView(AD_BIDS_NAME, undefined, CUSTOM_TEST_CONTROLS);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are persisted for custom
    assertMetricsView(AD_BIDS_MIRROR_NAME, undefined, CUSTOM_TEST_CONTROLS);
  });

  it("Select different measure", () => {
    createAdBidsInStore();
    assertMetricsView(AD_BIDS_NAME);

    // select a different leaderboard measure
    metricsExplorerStore.setLeaderboardMeasureName(
      AD_BIDS_NAME,
      AD_BIDS_BID_PRICE_MEASURE
    );
    assertMetricsView(
      AD_BIDS_NAME,
      undefined,
      undefined,
      AD_BIDS_BID_PRICE_MEASURE
    );

    // create a mirror using the proto and assert that the leaderboard measure is persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      ALL_TIME_PARSED_TEST_CONTROLS,
      AD_BIDS_BID_PRICE_MEASURE
    );
  });

  describe("Restore invalid state", () => {
    it("Restore invalid filter", () => {
      createAdBidsInStore();
      metricsExplorerStore.toggleFilter(
        AD_BIDS_NAME,
        AD_BIDS_PUBLISHER_DIMENSION,
        "Facebook"
      );
      metricsExplorerStore.toggleFilter(
        AD_BIDS_NAME,
        AD_BIDS_DOMAIN_DIMENSION,
        "google.com"
      );

      // create a mirror from state
      createAdBidsMirrorInStore();
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(
        AD_BIDS_MIRROR_NAME,
        AD_BIDS_WITH_DELETED_DIMENSION
      );
      // assert that the filter for removed dimension is not present anymore
      assertMetricsView(
        AD_BIDS_MIRROR_NAME,
        {
          include: [
            {
              name: AD_BIDS_PUBLISHER_DIMENSION,
              in: ["Facebook"],
            },
          ],
          exclude: [],
        },
        ALL_TIME_PARSED_TEST_CONTROLS
      );
    });

    it("Restore invalid leaderboard measure", () => {
      createAdBidsInStore();
      metricsExplorerStore.setLeaderboardMeasureName(
        AD_BIDS_NAME,
        AD_BIDS_BID_PRICE_MEASURE
      );

      // create a mirror from state
      createAdBidsMirrorInStore();
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(AD_BIDS_MIRROR_NAME, {
        name: "AdBids",
        measures: [
          {
            name: AD_BIDS_IMPRESSIONS_MEASURE,
            expression: "count(*)",
          },
        ],
        dimensions: [
          {
            name: AD_BIDS_PUBLISHER_DIMENSION,
          },
        ],
      });
      // assert that the selected measure is reset to the 1st available one
      expect(
        get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .leaderboardMeasureName
      ).toBe(AD_BIDS_IMPRESSIONS_MEASURE);
      expect([
        ...get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .visibleMeasureKeys,
      ]).toEqual([AD_BIDS_IMPRESSIONS_MEASURE]);
    });

    it("Restore invalid selected dimension", () => {
      createAdBidsInStore();
      metricsExplorerStore.setMetricDimensionName(
        AD_BIDS_NAME,
        AD_BIDS_DOMAIN_DIMENSION
      );

      // create a mirror from state
      createAdBidsMirrorInStore();
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(
        AD_BIDS_MIRROR_NAME,
        AD_BIDS_WITH_DELETED_DIMENSION
      );
      // assert that the selected dimension is cleared
      expect(
        get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .selectedDimensionName
      ).toBeUndefined();
      expect([
        ...get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .visibleDimensionKeys,
      ]).toEqual([AD_BIDS_PUBLISHER_DIMENSION]);
    });
  });
});
