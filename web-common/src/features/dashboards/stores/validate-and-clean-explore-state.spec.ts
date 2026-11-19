import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { validateAndCleanExploreState } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
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
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const publisherChip: PivotChipData = {
  id: AD_BIDS_PUBLISHER_DIMENSION,
  title: AD_BIDS_PUBLISHER_DIMENSION,
  type: PivotChipType.Dimension,
};
const staleDimensionChip: PivotChipData = {
  id: "stale_dimension",
  title: "stale_dimension",
  type: PivotChipType.Dimension,
};
const dayChip: PivotChipData = {
  id: "TIME_GRAIN_DAY",
  title: "day",
  type: PivotChipType.Time,
};
const impressionsChip: PivotChipData = {
  id: AD_BIDS_IMPRESSIONS_MEASURE,
  title: AD_BIDS_IMPRESSIONS_MEASURE,
  type: PivotChipType.Measure,
};
const staleMeasureChip: PivotChipData = {
  id: "stale_measure",
  title: "stale_measure",
  type: PivotChipType.Measure,
};
function getPivotState(
  rows: PivotChipData[],
  columns: PivotChipData[],
): PivotState {
  return {
    rows,
    columns,
    sorting: [],
    expanded: {},
    columnPage: 1,
    rowPage: 1,
    enableComparison: true,
    activeCell: null,
    showTotalsColumn: true,
    showTotalsRow: true,
    totalsRowPosition: "top",
    tableMode: "nest",
  };
}

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
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
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
  {
    title: "Filter on a stale dimension and a stale measure",
    exploreState: {
      whereFilter: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
        createInExpression("stale_dimension", ["Google"]),
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          ["stale_measure"],
          createBinaryExpression("stale_measure", V1Operation.OPERATION_GT, 10),
        ),
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          [AD_BIDS_BID_PRICE_MEASURE],
          createBinaryExpression(
            AD_BIDS_BID_PRICE_MEASURE,
            V1Operation.OPERATION_GT,
            10,
          ),
        ),
      ]),
    },
    expectedState: {
      whereFilter: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          [AD_BIDS_BID_PRICE_MEASURE],
          createBinaryExpression(
            AD_BIDS_BID_PRICE_MEASURE,
            V1Operation.OPERATION_GT,
            10,
          ),
        ),
      ]),
    },
    expectedErrorMessages: [
      `Selected filter dimension: "stale_dimension" is not valid.`,
      `Selected filter field: "stale_measure" is not valid.`,
    ],
  },
  {
    title: "Measure filter selecting a stale measure and a valid one",
    exploreState: {
      whereFilter: createAndExpression([
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          [AD_BIDS_IMPRESSIONS_MEASURE, "stale_measure"],
          createAndExpression([
            createBinaryExpression(
              AD_BIDS_IMPRESSIONS_MEASURE,
              V1Operation.OPERATION_GT,
              10,
            ),
            createBinaryExpression(
              "stale_measure",
              V1Operation.OPERATION_GT,
              5,
            ),
          ]),
        ),
      ]),
    },
    expectedState: {
      // The runtime resolves every measure the subquery selects, so the stale one is dropped from the list too.
      whereFilter: createAndExpression([
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          [AD_BIDS_IMPRESSIONS_MEASURE],
          createAndExpression([
            createBinaryExpression(
              AD_BIDS_IMPRESSIONS_MEASURE,
              V1Operation.OPERATION_GT,
              10,
            ),
          ]),
        ),
      ]),
    },
    expectedErrorMessages: [
      `Selected filter field: "stale_measure" is not valid.`,
    ],
  },
  {
    title: "Stale comparison, expanded and pivoted fields",
    exploreState: {
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimensionName: "stale_dimension",
      selectedComparisonDimension: "stale_dimension",
      pivot: getPivotState(
        [publisherChip, staleDimensionChip],
        [dayChip, staleMeasureChip, impressionsChip],
      ),
    },
    expectedState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      pivot: getPivotState([publisherChip], [dayChip, impressionsChip]),
    },
    expectedErrorMessages: [
      `Selected compare dimension: "stale_dimension" is not valid.`,
      `Selected expanded dimension: "stale_dimension" is not valid.`,
      `Selected pivot row: "stale_dimension" is not valid.`,
      `Selected pivot column: "stale_measure" is not valid.`,
    ],
  },
  {
    title: "Stale expanded measure",
    exploreState: {
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      tdd: {
        expandedMeasureName: "stale_measure",
        chartType: TDDChart.DEFAULT,
        pinIndex: -1,
      },
    },
    expectedState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      tdd: {
        expandedMeasureName: "",
        chartType: TDDChart.DEFAULT,
        pinIndex: -1,
      },
    },
    expectedErrorMessages: [
      `Selected expanded measure: "stale_measure" is not valid.`,
    ],
  },
  {
    title:
      "Leaderboard sync invariant - sort measure not in leaderboard measures",
    exploreState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
    },
    expectedState: {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
    },
    expectedErrorMessages: [],
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
