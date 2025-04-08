import { getBlankExploreState } from "@rilldata/web-common/features/dashboards/stores/get-blank-explore-state";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlSearch } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-search";
import { convertUrlSearchToPartialExploreState } from "@rilldata/web-common/features/dashboards/url-state/convert-url-search-to-partial-explore-state";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const TestCases: {
  title: string;
  partialExploreState: Partial<MetricsExplorerEntity>;
  expectedUrlSearch: string;
  expectedPartialExploreState?: Partial<MetricsExplorerEntity>;
}[] = [
  {
    title: "Same time settings as blank explore",
    partialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      selectedTimeRange: {
        name: TimeRangePreset.LAST_SIX_HOURS,
        interval: V1TimeGrain.TIME_GRAIN_HOUR,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,
    },
    expectedUrlSearch: "view=explore",
    expectedPartialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
    },
  },
  {
    title: "Different time settings as blank explore",
    partialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      selectedTimeRange: {
        name: TimeRangePreset.LAST_7_DAYS,
        interval: V1TimeGrain.TIME_GRAIN_DAY,
      } as DashboardTimeControls,
      showTimeComparison: true,
      selectedComparisonTimeRange: {
        name: TimeComparisonOption.DAY,
      } as DashboardTimeControls,
      selectedTimezone: "Asia/Kolkata",
      selectedScrubRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-01-01T00:00:00.000Z"),
        end: new Date("2025-02-01T00:00:00.000Z"),
        isScrubbing: false,
      },
    },
    expectedUrlSearch:
      "view=explore&tr=P7D&tz=Asia%2FKolkata&compare_tr=rill-PD&grain=day&highlighted_tr=2025-01-01T00%3A00%3A00.000Z%2C2025-02-01T00%3A00%3A00.000Z",
  },

  {
    title: "Same leaderboard settings as blank dashboard",
    partialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      visibleMeasures: [
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
        AD_BIDS_PUBLISHER_COUNT_MEASURE,
      ],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_COUNTRY_DIMENSION,
      ],
      allDimensionsVisible: true,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      leaderboardMeasureCount: 1,
    },
    expectedUrlSearch: "view=explore",
    expectedPartialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
    },
  },
  {
    title: "Different leaderboard settings as blank dashboard",
    partialExploreState: {
      activePage: DashboardState_ActivePage.DEFAULT,
      visibleMeasures: [
        AD_BIDS_BID_PRICE_MEASURE,
        AD_BIDS_PUBLISHER_COUNT_MEASURE,
      ],
      allMeasuresVisible: false,
      visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION, AD_BIDS_COUNTRY_DIMENSION],
      allDimensionsVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
      leaderboardMeasureCount: 2,
    },
    expectedUrlSearch:
      "view=explore&measures=bid_price%2Cpublisher_count&dims=domain%2Ccountry&sort_by=bid_price&sort_type=percent&sort_dir=ASC&leaderboard_measure_count=2",
  },
];

const TestMetricsViewSpec = AD_BIDS_METRICS_3_MEASURES_DIMENSIONS;
const TestExploreSpec = AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS;

describe("partial explore state <==> url search", () => {
  it("Should create a blank explore url params", () => {
    const blankExploreUrlParams = getBlankExploreUrlParams();
    expect(blankExploreUrlParams.toString()).toEqual(
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&f=&measures=*&dims=*&expand_dim=&sort_by=impressions&sort_type=value&sort_dir=DESC&leaderboard_measure_count=1",
    );
  });

  for (const {
    title,
    partialExploreState,
    expectedUrlSearch,
    expectedPartialExploreState,
  } of TestCases) {
    it(title, () => {
      const timeControlState = getTimeControlState(
        TestMetricsViewSpec,
        TestExploreSpec,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        partialExploreState as any,
      );

      // Convert to url using the blankExploreUrlParams
      const urlParamsUsingBlankParams = convertPartialExploreStateToUrlSearch(
        partialExploreState,
        TestExploreSpec,
        timeControlState,
        getBlankExploreUrlParams(),
      );
      expect(urlParamsUsingBlankParams.toString()).toEqual(expectedUrlSearch);

      const { partialExploreState: partialExploreStateUsingBlankParams } =
        convertUrlSearchToPartialExploreState(
          urlParamsUsingBlankParams,
          TestMetricsViewSpec,
          TestExploreSpec,
        );
      expect(partialExploreStateUsingBlankParams).toEqual(
        expectedPartialExploreState ?? partialExploreState,
      );

      // Converting to url and back without passing blankExploreUrlParams should get the exact input partial explore state
      const urlParamsNotUsingBlankParams =
        convertPartialExploreStateToUrlSearch(
          partialExploreState,
          TestExploreSpec,
          timeControlState,
          new URLSearchParams(),
        );
      const { partialExploreState: partialExploreStateNotUsingBlankParams } =
        convertUrlSearchToPartialExploreState(
          urlParamsNotUsingBlankParams,
          TestMetricsViewSpec,
          TestExploreSpec,
        );

      expect(partialExploreStateNotUsingBlankParams).toEqual(
        partialExploreState,
      );
    });
  }
});

export function getBlankExploreUrlParams() {
  const blankExploreState = getBlankExploreState(
    TestMetricsViewSpec,
    TestExploreSpec,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
  const timeControlState = getTimeControlState(
    TestMetricsViewSpec,
    TestExploreSpec,
    AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    blankExploreState as any,
  );
  return convertPartialExploreStateToUrlSearch(
    blankExploreState,
    TestExploreSpec,
    timeControlState,
    new URLSearchParams(),
  );
}
