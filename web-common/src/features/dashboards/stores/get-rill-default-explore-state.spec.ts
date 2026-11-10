import {
  getDefaultTimeZone,
  getRillDefaultExploreState,
} from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
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

describe("getDefaultTimeZone", () => {
  it("uses defaultPreset.timezone over timeZones[0]", () => {
    const explore: V1ExploreSpec = {
      defaultPreset: { timezone: "Asia/Shanghai" },
      timeZones: ["UTC", "Asia/Shanghai"],
    };
    expect(getDefaultTimeZone(explore)).toBe("Asia/Shanghai");
  });

  it("uses the first explore time zone when no preset timezone is set", () => {
    const explore: V1ExploreSpec = {
      timeZones: ["America/Los_Angeles", "UTC"],
    };
    expect(getDefaultTimeZone(explore)).toBe("America/Los_Angeles");
  });

  it("falls back to DEFAULT_TIMEZONES[0] when the explore lists none", () => {
    expect(getDefaultTimeZone({})).toBe(DEFAULT_TIMEZONES[0]);
    expect(DEFAULT_TIMEZONES[0]).toBe("UTC");
  });

  it("resolves Local to the browser IANA zone", () => {
    const explore: V1ExploreSpec = { timeZones: ["Local"] };
    expect(getDefaultTimeZone(explore)).toBe(getLocalIANA());
  });

  it("falls back to UTC for an invalid IANA zone", () => {
    const explore: V1ExploreSpec = { timeZones: ["Not/A_Timezone"] };
    expect(getDefaultTimeZone(explore)).toBe("UTC");
  });
});
