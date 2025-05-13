import { validateAndCleanExploreState } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { describe, expect, it } from "vitest";

const TestCases: {
  title: string;
  exploreState: Partial<ExploreState>;
  expectedState: Partial<ExploreState>;
  expectedErrorMessages: string[];
}[] = [
  {
    title: "Some invalid selected measures/dimensions",
    exploreState: {
      visibleMeasures: [
        AD_BIDS_IMPRESSIONS_MEASURE,
        "invalid_measure",
        AD_BIDS_BID_PRICE_MEASURE,
      ],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        "invalid_dimension",
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: true,
    },
    expectedState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: false,
    },
    expectedErrorMessages: [
      `Selected dimension: "invalid_dimension" is not valid.`,
      `Selected measure: "invalid_measure" is not valid.`,
    ],
  },
  {
    title: "All invalid selected measures/dimensions",
    exploreState: {
      visibleMeasures: [
        "invalid_measure_1",
        "invalid_measure_2",
        "invalid_measure_3",
      ],
      allMeasuresVisible: true,
      visibleDimensions: [
        "invalid_dimension_1",
        "invalid_dimension_2",
        "invalid_dimension_3",
      ],
      allDimensionsVisible: true,
      leaderboardSortByMeasureName: "invalid_measure_1",
      leaderboardMeasureNames: ["invalid_measure_1", "invalid_measure_2"],
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    },
    expectedState: {
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    },
    expectedErrorMessages: [
      `Selected dimensions: "invalid_dimension_1,invalid_dimension_2,invalid_dimension_3" are not valid.`,
      `Selected measures: "invalid_measure_1,invalid_measure_2,invalid_measure_3" are not valid.`,
    ],
  },

  {
    title: "Hidden sort settings",
    exploreState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_PUBLISHER_COUNT_MEASURE,
      leaderboardMeasureNames: [
        AD_BIDS_PUBLISHER_COUNT_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ],
    },
    expectedState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
    },
    expectedErrorMessages: [
      `Selected sort by measure: "publisher_count" is not valid. It is hidden.`,
      `Selected leaderboard measure: "publisher_count" is not valid.`,
    ],
  },
  {
    title: "All invalid sort settings",
    exploreState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: "invalid_measure_1",
      leaderboardMeasureNames: ["invalid_measure_1", "invalid_measure_2"],
    },
    expectedState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
    },
    expectedErrorMessages: [
      `Selected sort by measure: "invalid_measure_1" is not valid.`,
      `Selected leaderboard measures: "invalid_measure_1,invalid_measure_2" are not valid.`,
    ],
  },
];

describe("validateAndCleanExploreState", () => {
  for (const {
    title,
    exploreState,
    expectedState,
    expectedErrorMessages,
  } of TestCases) {
    it(title, () => {
      const errors = validateAndCleanExploreState(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
        exploreState,
      );

      expect(exploreState).toEqual(expectedState);
      expect(errors.map((e) => e.message)).toEqual(expectedErrorMessages);
    });
  }
});
