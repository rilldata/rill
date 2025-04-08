import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
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
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const blankExploreUrlParams = new URLSearchParams(
  "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&f=&measures=*&dims=*&sort_by=impressions&sort_type=value&sort_dir=ASC&leaderboard_measure_count=1",
);
const TestCases: {
  title: string;
  partialExploreState: Partial<MetricsExplorerEntity>;
  expectedUrlSearch: string;
  expectedPartialExploreState?: Partial<MetricsExplorerEntity>;
}[] = [
  {
    title: "Explore state same time settings",
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
    title: "Explore state different time settings",
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
    },
    expectedUrlSearch: "view=explore&tr=P7D&compare_tr=rill-PD&grain=day",
  },
];

describe("partial explore state <==> url search", () => {
  for (const {
    title,
    partialExploreState,
    expectedUrlSearch,
    expectedPartialExploreState,
  } of TestCases) {
    it(title, () => {
      const timeControlState = getTimeControlState(
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        partialExploreState as any,
      );

      // Convert to url using the blankExploreUrlParams
      const urlParamsUsingBlankParams = convertPartialExploreStateToUrlSearch(
        partialExploreState,
        AD_BIDS_EXPLORE_INIT,
        timeControlState,
        blankExploreUrlParams,
      );
      expect(urlParamsUsingBlankParams.toString()).toEqual(expectedUrlSearch);

      const { partialExploreState: partialExploreStateUsingBlankParams } =
        convertUrlSearchToPartialExploreState(
          urlParamsUsingBlankParams,
          AD_BIDS_METRICS_INIT_WITH_TIME,
          AD_BIDS_EXPLORE_INIT,
        );
      expect(partialExploreStateUsingBlankParams).toEqual(
        expectedPartialExploreState ?? partialExploreState,
      );

      // Converting to url and back without passing blankExploreUrlParams should get the exact input partial explore state
      const urlParamsNotUsingBlankParams =
        convertPartialExploreStateToUrlSearch(
          partialExploreState,
          AD_BIDS_EXPLORE_INIT,
          timeControlState,
          new URLSearchParams(),
        );
      const { partialExploreState: partialExploreStateNotUsingBlankParams } =
        convertUrlSearchToPartialExploreState(
          urlParamsNotUsingBlankParams,
          AD_BIDS_METRICS_INIT_WITH_TIME,
          AD_BIDS_EXPLORE_INIT,
        );

      expect(partialExploreStateNotUsingBlankParams).toEqual(
        partialExploreState,
      );
    });
  }
});
