import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import {
  getMostRecentExploreUrlParams,
  saveMostRecentExploreUrlParams,
  setLegacyMostRecentExploreStateInLocalStorage,
  setMostRecentExploreUrlParamsInLocalStorage,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { beforeEach, describe, expect, it } from "vitest";

const doubled: EphemeralMeasureDef = {
  name: "doubled",
  displayName: "Doubled",
  expression: "impressions * 2",
};

// Each case saves the same state as params and as the legacy JSON state; both must load the same.
const TestCases: {
  title: string;
  params: string;
  legacyState: Partial<ExploreState>;
  ephemeralMeasures?: EphemeralMeasureDef[];
  expected: string;
}[] = [
  {
    title: "keeps valid fields",
    params:
      "tz=Asia/Kathmandu&measures=impressions&dims=domain&sort_by=impressions&sort_dir=ASC&leaderboard_measures=impressions",
    legacyState: {
      selectedTimezone: "Asia/Kathmandu",
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
    },
    expected:
      "tz=Asia/Kathmandu&measures=impressions&dims=domain&sort_by=impressions&sort_dir=ASC&leaderboard_measures=impressions",
  },
  {
    title: "drops visible fields that are all stale, along with the sort",
    params:
      "measures=stale_measure&dims=stale_dimension&sort_by=stale_measure&leaderboard_measures=stale_measure",
    legacyState: {
      visibleMeasures: ["stale_measure"],
      visibleDimensions: ["stale_dimension"],
      leaderboardSortByMeasureName: "stale_measure",
      leaderboardMeasureNames: ["stale_measure"],
    },
    expected: "",
  },
  {
    title: "falls back to the first visible measure for a stale sort",
    params: "measures=bid_price,impressions&sort_by=stale_measure",
    legacyState: {
      visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE, AD_BIDS_IMPRESSIONS_MEASURE],
      leaderboardSortByMeasureName: "stale_measure",
    },
    expected: "measures=bid_price,impressions&sort_by=bid_price",
  },
  {
    title: "keeps the valid leaderboard measures",
    params:
      "measures=impressions,bid_price&leaderboard_measures=stale_measure,bid_price",
    legacyState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      leaderboardMeasureNames: ["stale_measure", AD_BIDS_BID_PRICE_MEASURE],
    },
    // The visible measures are all of the explore's.
    expected: "measures=*&leaderboard_measures=bid_price",
  },
  {
    title: "drops ad-hoc names when their definitions are unknown",
    params:
      "measures=impressions,doubled&sort_by=doubled&leaderboard_measures=doubled",
    legacyState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, doubled.name],
      leaderboardSortByMeasureName: doubled.name,
      leaderboardMeasureNames: [doubled.name],
    },
    expected:
      "measures=impressions&sort_by=impressions&leaderboard_measures=impressions",
  },
  {
    title: "keeps ad-hoc names when the library's definitions are provided",
    params:
      "measures=impressions,doubled&sort_by=doubled&leaderboard_measures=doubled",
    legacyState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, doubled.name],
      leaderboardSortByMeasureName: doubled.name,
      leaderboardMeasureNames: [doubled.name],
    },
    ephemeralMeasures: [doubled],
    // The definitions themselves stay in the library.
    expected:
      "measures=impressions,doubled&sort_by=doubled&leaderboard_measures=doubled",
  },
];

describe("getMostRecentExploreUrlParams", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  for (const {
    title,
    params,
    legacyState,
    ephemeralMeasures,
    expected,
  } of TestCases) {
    it(`${title}, from params`, () => {
      setMostRecentExploreUrlParamsInLocalStorage(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        new URLSearchParams(params),
      );
      assertMostRecentParams(expected, ephemeralMeasures);
    });

    it(`${title}, from the legacy state`, () => {
      setLegacyMostRecentExploreStateInLocalStorage(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        legacyState,
      );
      assertMostRecentParams(expected, ephemeralMeasures);
    });
  }

  it("prefers params over the legacy state", () => {
    setLegacyMostRecentExploreStateInLocalStorage(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      { visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE] },
    );
    setMostRecentExploreUrlParamsInLocalStorage(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      new URLSearchParams("measures=bid_price"),
    );
    assertMostRecentParams("measures=bid_price");
  });
});

describe("saveMostRecentExploreUrlParams", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("saves the long term subset of the explore view's params", () => {
    saveMostRecentExploreUrlParams(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      new URLSearchParams(
        "view=explore&tr=P7D&tz=UTC&f=publisher IN ('Google')&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=impressions&sort_type=percent&sort_dir=ASC&leaderboard_measures=impressions&lb_ctx=true&adhoc_m=doubled:Doubled:impressions%20*%202",
      ),
    );
    assertMostRecentParams(
      `tz=UTC&measures=impressions&dims=${AD_BIDS_PUBLISHER_DIMENSION}&sort_by=impressions&sort_type=percent&sort_dir=ASC&leaderboard_measures=impressions&lb_ctx=true`,
    );
  });

  it("does not save params of other views", () => {
    saveMostRecentExploreUrlParams(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      new URLSearchParams("view=pivot&tz=UTC&sort_by=publisher"),
    );
    expect(
      getMostRecentExploreUrlParams(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
      ),
    ).toBeUndefined();
  });
});

function assertMostRecentParams(
  expected: string,
  ephemeralMeasures?: EphemeralMeasureDef[],
) {
  const params = getMostRecentExploreUrlParams(
    AD_BIDS_EXPLORE_NAME,
    undefined,
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    AD_BIDS_EXPLORE_INIT,
    ephemeralMeasures,
  );
  expect(params?.toString()).toEqual(new URLSearchParams(expected).toString());
}
