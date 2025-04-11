import { getDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/stores/get-default-explore-url-params";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { V1ExploreSortType } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("getDefaultExploreUrlParams", () => {
  it("Metrics explore without a preset", () => {
    const defaultExploreUrlParams = getDefaultExploreUrlParams(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );
    expect(defaultExploreUrlParams.toString()).toEqual(
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=&f=&measures=*&dims=*&expand_dim=&sort_by=impressions&sort_type=value&sort_dir=DESC&leaderboard_measures=impressions",
    );
  });

  it("Metrics explore with a preset", () => {
    const defaultExploreUrlParams = getDefaultExploreUrlParams(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      {
        ...AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
        defaultPreset: {
          timeRange: "P7D",
          timeGrain: "day",
          timezone: "Asia/Kolkata",
          compareTimeRange: "rill-PW",

          measures: [
            AD_BIDS_BID_PRICE_MEASURE,
            AD_BIDS_PUBLISHER_COUNT_MEASURE,
          ],
          dimensions: [AD_BIDS_DOMAIN_DIMENSION, AD_BIDS_COUNTRY_DIMENSION],
          exploreSortBy: AD_BIDS_BID_PRICE_MEASURE,
          exploreSortAsc: true,
          exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_PERCENT,
          exploreLeaderboardMeasures: [
            AD_BIDS_BID_PRICE_MEASURE,
            AD_BIDS_PUBLISHER_COUNT_MEASURE,
          ],
        },
      },
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );
    expect(defaultExploreUrlParams.toString()).toEqual(
      "view=explore&tr=P7D&tz=Asia%2FKolkata&compare_tr=&grain=day&compare_dim=&f=&measures=bid_price%2Cpublisher_count&dims=domain%2Ccountry&expand_dim=&sort_by=bid_price&sort_type=percent&sort_dir=ASC&leaderboard_measures=bid_price%2Cpublisher_count",
    );
  });
});
