import { type CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { useDashboardFetchMocksForComponentTests } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import { setExploreStateForWebView } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { setMostRecentExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import DashboardStateManagerTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/DashboardStateManagerTest.svelte";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { render, screen, waitFor } from "@testing-library/svelte";
import { readable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
  };
});

// Contains basic tests for verifying order of selection.
// In-depth tests for storing and retrieving state are separate.
describe("DashboardStateManager", () => {
  mockAnimationsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForExploreTests;

  beforeEach(() => {
    pageMock = new PageMockForExploreTests(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: {
          ...AD_BIDS_PRESET,
          comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
        },
      },
    );
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, {
      min: "2024-01-01",
      max: "2024-03-31",
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  describe("Dashboards with timeseries", () => {
    const ExploreStateSubsetForRillDefaultState: Partial<MetricsExplorerEntity> =
      {
        selectedTimeRange: {
          name: "P7D",
          interval: V1TimeGrain.TIME_GRAIN_DAY,
        } as DashboardTimeControls,
        showTimeComparison: false,
        selectedComparisonTimeRange: undefined,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      };
    const PageURLForRillDefaultState =
      "view=explore&tr=P7D&tz=Asia%2FKathmandu&compare_tr=&grain=day&f=&measures=impressions&dims=publisher&expand_dim=&sort_by=impressions&sort_type=percent&sort_dir=ASC&leaderboard_measure_count=1";
    const BookmarkSourceQueryResult = readable({
      data: {
        selectedTimeRange: {
          name: "PT24H",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls,
      },
      error: null,
      isFetching: false,
      isLoading: false,
    });

    it("Should load base dashboard state", async () => {
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset(ExploreStateSubsetForRillDefaultState);
      expect(pageMock.urlSearchHistory).toEqual([PageURLForRillDefaultState]);
    });

    it("Should load 'other source' of dashboard state", async () => {
      renderDashboardStateManager(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,

        selectedTimeRange: {
          name: "PT24H",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls,
      });
      const initUrlSearch =
        "view=explore&tr=PT24H&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=hour&f=&measures=impressions&dims=publisher&expand_dim=&sort_by=impressions&sort_type=percent&sort_dir=ASC&leaderboard_measure_count=1";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        initUrlSearch,
        PageURLForRillDefaultState,
      ]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreState(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      renderDashboardStateManager(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,

        // Remaining settings from yaml defaults
        selectedTimeRange: {
          name: "PT24H",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls,
      });
      const initUrlSearch =
        "view=explore&tr=PT24H&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=hour&f=&measures=bid_price&dims=domain&expand_dim=&sort_by=bid_price&sort_type=value&sort_dir=ASC&leaderboard_measure_count=1";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        initUrlSearch,
        PageURLForRillDefaultState,
      ]);
    });

    it("Should load from session dashboard state", async () => {
      setExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        ExploreUrlWebView.Explore,
        "view=explore&tr=P14D&compare_tr=rill-PW&grain=day&measures=bid_price&dims=domain&sort_by=bid_price&sort_dir=DESC",
      );
      renderDashboardStateManager(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        selectedTimeRange: {
          name: "P14D",
          interval: V1TimeGrain.TIME_GRAIN_DAY,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.WEEK,
        } as DashboardTimeControls,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      });
      const initUrlSearch =
        "view=explore&tr=P14D&tz=Asia%2FKathmandu&compare_tr=rill-PW&grain=day&f=&measures=bid_price&dims=domain&expand_dim=&sort_by=bid_price&sort_type=percent&sort_dir=DESC&leaderboard_measure_count=1";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        initUrlSearch,
        PageURLForRillDefaultState,
      ]);
    });
  });

  describe("Dashboards without timeseries", () => {
    const ExploreStateSubsetForRillDefaultState: Partial<MetricsExplorerEntity> =
      {
        selectedTimeRange: undefined,
        showTimeComparison: false,
        selectedComparisonTimeRange: undefined,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      };
    const PageURLForRillDefaultState =
      "view=explore&f=&measures=impressions&dims=publisher&expand_dim=&sort_by=impressions&sort_type=percent&sort_dir=ASC&leaderboard_measure_count=1";

    beforeEach(() => {
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
      });
    });

    it("Should load base dashboard state", async () => {
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset(ExploreStateSubsetForRillDefaultState);
      expect(pageMock.urlSearchHistory).toEqual([PageURLForRillDefaultState]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreState(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      const initUrlSearch =
        "view=explore&f=&measures=bid_price&dims=domain&expand_dim=&sort_by=bid_price&sort_type=value&sort_dir=ASC&leaderboard_measure_count=1";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        initUrlSearch,
        PageURLForRillDefaultState,
      ]);
    });

    it("Should load from session dashboard state", async () => {
      setExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        ExploreUrlWebView.Explore,
        "view=explore&measures=bid_price&dims=domain&sort_by=bid_price&sort_dir=DESC",
      );
      renderDashboardStateManager();

      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      assertExploreStateSubset({
        selectedTimeRange: undefined,
        showTimeComparison: false,
        selectedComparisonTimeRange: undefined,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      });
      const initUrlSearch =
        "view=explore&f=&measures=bid_price&dims=domain&expand_dim=&sort_by=bid_price&sort_type=percent&sort_dir=DESC&leaderboard_measure_count=1";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        initUrlSearch,
        PageURLForRillDefaultState,
      ]);
    });
  });
});

// This needs to be there each file because of how hoisting works with vitest.
// TODO: find if there is a way to share code.
function renderDashboardStateManager(
  bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<MetricsExplorerEntity> | undefined>
    | undefined = undefined,
) {
  const renderResults = render(DashboardStateManagerTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
      bookmarkOrTokenExploreState,
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map([["$$_queryClient", queryClient]]),
  });

  return { queryClient, renderResults };
}

function assertExploreStateSubset(
  exploreStateSubset: Partial<MetricsExplorerEntity>,
) {
  const curExploreState = getCleanMetricsExploreForAssertion();
  const curExploreStateSubset: Partial<MetricsExplorerEntity> = {
    selectedTimeRange: curExploreState.selectedTimeRange,
    showTimeComparison: curExploreState.showTimeComparison,
    selectedComparisonTimeRange: curExploreState.selectedComparisonTimeRange,

    visibleMeasures: curExploreState.visibleMeasures,
    allMeasuresVisible: curExploreState.allMeasuresVisible,

    visibleDimensions: curExploreState.visibleDimensions,
    allDimensionsVisible: curExploreState.allDimensionsVisible,

    leaderboardSortByMeasureName: curExploreState.leaderboardSortByMeasureName,
    dashboardSortType: curExploreState.dashboardSortType,
    sortDirection: curExploreState.sortDirection,
  };
  expect(curExploreStateSubset).toEqual(exploreStateSubset);
}
