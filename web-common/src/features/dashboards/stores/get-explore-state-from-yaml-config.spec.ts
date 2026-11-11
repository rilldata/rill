import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

// An explore spec that still names fields the metrics view dropped.
const StaleExploreSpec: V1ExploreSpec = {
  ...AD_BIDS_EXPLORE_INIT,
  measures: [
    AD_BIDS_IMPRESSIONS_MEASURE,
    AD_BIDS_BID_PRICE_MEASURE,
    "stale_measure",
  ],
  dimensions: [
    AD_BIDS_PUBLISHER_DIMENSION,
    AD_BIDS_DOMAIN_DIMENSION,
    "stale_dimension",
  ],
};

describe("getExploreStateFromYAMLConfig", () => {
  it("drops the defaults the metrics view no longer defines", () => {
    const exploreState = getExploreStateFromYAMLConfig(
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...StaleExploreSpec,
        defaultPreset: {
          measures: ["stale_measure", AD_BIDS_BID_PRICE_MEASURE],
          dimensions: ["stale_dimension", AD_BIDS_DOMAIN_DIMENSION],
          exploreSortBy: "stale_measure",
          exploreLeaderboardMeasures: ["stale_measure"],
          exploreExpandedDimension: "stale_dimension",
          comparisonMode:
            V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION,
          comparisonDimension: "stale_dimension",
        },
      },
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );

    expect(exploreState).toEqual({
      activePage: DashboardState_ActivePage.DEFAULT,
      // Falls back to the first dimension that still exists.
      selectedComparisonDimension: AD_BIDS_PUBLISHER_DIMENSION,
      visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
      allDimensionsVisible: false,
      // Fall back to the first visible measure.
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
    });
  });

  it("leaves fields unset when no default field exists any more", () => {
    const exploreState = getExploreStateFromYAMLConfig(
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...StaleExploreSpec,
        defaultPreset: {
          measures: ["stale_measure"],
          dimensions: ["stale_dimension"],
          exploreSortBy: "stale_measure",
          exploreLeaderboardMeasures: ["stale_measure"],
        },
      },
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );

    // The cascading merge fills these from the next source.
    expect(exploreState).toEqual({
      activePage: DashboardState_ActivePage.DEFAULT,
    });
  });
});
