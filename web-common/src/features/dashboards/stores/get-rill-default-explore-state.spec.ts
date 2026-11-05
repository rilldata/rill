import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { describe, expect, it } from "vitest";

describe("getRillDefaultExploreState", () => {
  it("only uses the explore fields the metrics view still defines", () => {
    const exploreState = getRillDefaultExploreState(
      AD_BIDS_METRICS_INIT,
      {
        ...AD_BIDS_EXPLORE_INIT,
        measures: [
          "stale_measure",
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],
        dimensions: [
          AD_BIDS_PUBLISHER_DIMENSION,
          "stale_dimension",
          AD_BIDS_DOMAIN_DIMENSION,
        ],
      },
      undefined,
    );

    expect(exploreState).toMatchObject({
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: true,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
    });
  });
});
