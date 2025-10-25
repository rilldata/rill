import { exploreBookmarkDataTransformer } from "@rilldata/web-admin/features/bookmarks/explore-bookmark-legacy-data-transformer.ts";
import {
  getBookmarkData,
  parseBookmarks,
} from "@rilldata/web-admin/features/bookmarks/utils.ts";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto.ts";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_RILL_DEFAULT_EXPLORE_STATE,
  AD_BIDS_RILL_DEFAULT_EXPLORE_URL_PARAMS,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, it, expect, vi, beforeEach } from "vitest";

const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
  };
});

describe("getBookmarkData and parseBookmarks", () => {
  beforeEach(() => {
    new PageMockForExploreTests(hoistedPage);
  });

  const TestCases: {
    title: string;
    filtersOnly: boolean;
    partialExploreState: Partial<ExploreState>;
    urls: {
      subTitle: string;
      curUrlSearch: string;
      expectedFullUrlSearch: string;
      isActive: boolean;
    }[];
  }[] = [
    {
      title: "Complete bookmark",
      filtersOnly: false,
      partialExploreState: {
        activePage: DashboardState_ActivePage.DIMENSION_TABLE,
        whereFilter: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Yahoo",
          ]),
        ]),
        selectedTimeRange: {
          name: "P7D",
          interval: V1TimeGrain.TIME_GRAIN_HOUR, // Equal to default
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: "rill-PP",
        } as DashboardTimeControls,
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
      },
      urls: [
        {
          subTitle: "Empty url",
          curUrlSearch: "",
          expectedFullUrlSearch:
            "view=explore&tr=P7D&tz=UTC&grain=hour&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&compare_dim=&measures=bid_price&dims=*&expand_dim=domain&sort_by=bid_price&sort_dir=DESC&sort_type=value&leaderboard_measures=bid_price",
          isActive: false,
        },
        {
          subTitle: "Filter only equal",
          curUrlSearch:
            "view=tdd&tr=P7D&compare_tr=rill-PW&f=publisher+IN+('Facebook','Yahoo')&measure=impressions&chart_type=stacked_bar",
          expectedFullUrlSearch:
            "view=explore&tr=P7D&tz=UTC&grain=hour&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&compare_dim=&measures=bid_price&dims=*&expand_dim=domain&sort_by=bid_price&sort_dir=DESC&sort_type=value&leaderboard_measures=bid_price",
          isActive: false,
        },
        {
          subTitle: "Same url",
          curUrlSearch:
            "tr=P7D&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&measures=bid_price&expand_dim=domain&sort_by=bid_price&leaderboard_measures=bid_price",
          expectedFullUrlSearch:
            "view=explore&tr=P7D&tz=UTC&grain=hour&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&compare_dim=&measures=bid_price&dims=*&expand_dim=domain&sort_by=bid_price&sort_dir=DESC&sort_type=value&leaderboard_measures=bid_price",
          isActive: true,
        },
      ],
    },

    {
      title: "Filter only bookmark",
      filtersOnly: true,
      partialExploreState: {
        whereFilter: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Yahoo",
          ]),
        ]),
        selectedTimeRange: {
          name: "P7D",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
      },
      urls: [
        {
          subTitle: "Empty url",
          curUrlSearch: "",
          expectedFullUrlSearch:
            "f=publisher+IN+('Facebook','Yahoo')&grain=hour&tr=P7D",
          isActive: false,
        },
        {
          subTitle: "Different filter url",
          curUrlSearch: "tr=P2D",
          expectedFullUrlSearch:
            "f=publisher+IN+('Facebook','Yahoo')&grain=hour&tr=P7D",
          isActive: false,
        },
        {
          subTitle: "Filter only equal",
          curUrlSearch:
            "view=tdd&tr=P7D&compare_tr=rill-PW&f=publisher+IN+('Facebook','Yahoo')&measure=impressions&chart_type=stacked_bar",
          expectedFullUrlSearch:
            "view=tdd&tr=P7D&grain=hour&compare_tr=rill-PW&f=publisher+IN+('Facebook','Yahoo')&measure=impressions&chart_type=stacked_bar",
          isActive: true,
        },
        {
          subTitle: "Same url",
          curUrlSearch:
            "tr=P7D&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&measures=bid_price&expand_dim=domain&sort_by=bid_price&leaderboard_measures=bid_price",
          expectedFullUrlSearch:
            "tr=P7D&grain=hour&compare_tr=rill-PP&f=publisher+IN+('Facebook','Yahoo')&measures=bid_price&expand_dim=domain&sort_by=bid_price&leaderboard_measures=bid_price",
          isActive: true,
        },
      ],
    },
  ];

  for (const { title, filtersOnly, partialExploreState, urls } of TestCases) {
    const fullExploreState = {
      ...AD_BIDS_RILL_DEFAULT_EXPLORE_STATE,
      ...partialExploreState,
    };
    // Generate the bookmark url from the default explore state + modified partial explore state
    const curUrlParams = convertPartialExploreStateToUrlParams(
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
      fullExploreState,
      getTimeControlState(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
        AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        fullExploreState,
      ),
    );
    // Generate the bookmark data as base64 of the url search
    const bookmarkUrlData = getBookmarkData({
      curUrlParams,
      defaultUrlParams: AD_BIDS_RILL_DEFAULT_EXPLORE_URL_PARAMS,
      filtersOnly,
    });
    // Generate the bookmark data as proto state for legacy parsing test
    const exploreStateForProtoBookmark = filtersOnly
      ? ({
          whereFilter: fullExploreState.whereFilter,
          dimensionThresholdFilters: fullExploreState.dimensionThresholdFilters,
          selectedTimeRange: fullExploreState.selectedTimeRange,
        } as ExploreState)
      : fullExploreState;
    const bookmarkProtoData = getProtoFromDashboardState(
      exploreStateForProtoBookmark,
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
    );

    describe(title, () => {
      for (const {
        subTitle,
        curUrlSearch,
        expectedFullUrlSearch,
        isActive,
      } of urls) {
        it(subTitle, () => {
          // Parse both new and old formats together
          const [parsedBookmark, parsedProtoBookmark] = parseBookmarks(
            [
              { urlSearch: bookmarkUrlData, displayName: "new" },
              { data: bookmarkProtoData, displayName: "old" },
            ],
            new URLSearchParams(curUrlSearch),
            AD_BIDS_RILL_DEFAULT_EXPLORE_URL_PARAMS,
            (data) =>
              // Use the explore transformer to match the ExploreBookmarks components and to parse legacy proto state
              exploreBookmarkDataTransformer({
                data,
                metricsViewSpec:
                  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
                exploreSpec: AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
                timeRangeSummary: AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
              }),
          );

          assertUnorderedUrlSearch(
            parsedBookmark.fullUrl.slice(1),
            expectedFullUrlSearch,
          );
          expect(parsedBookmark.isActive).toEqual(isActive);

          assertUnorderedUrlSearch(parsedProtoBookmark.url, parsedBookmark.url);
          assertUnorderedUrlSearch(
            parsedProtoBookmark.fullUrl,
            parsedBookmark.fullUrl,
          );
          expect(parsedProtoBookmark.isActive).toEqual(isActive);
        });
      }
    });
  }
});

function assertUnorderedUrlSearch(actual: string, expected: string) {
  const actualUrlParams = new URLSearchParams(actual);
  actualUrlParams.sort();

  const expectedUrlSearchParams = new URLSearchParams(expected);
  expectedUrlSearchParams.sort();

  expect(decodeURIComponent(actualUrlParams.toString())).toEqual(
    decodeURIComponent(expectedUrlSearchParams.toString()),
  );
}
