import { type CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { useDashboardFetchMocksForComponentTests } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import { setMostRecentExploreStateForWebView } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-store";
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
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  type HoistedPage,
  PageMock,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMock";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { render, screen, waitFor } from "@testing-library/svelte";
import { readable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DashboardStateManagerTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/DashboardStateManagerTest.svelte";

const hoistedPage: HoistedPage = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url) => hoistedPage.goto(url),
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
  let pageMock!: PageMock;

  beforeEach(() => {
    pageMock = new PageMock(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, {
      min: "2024-01-01",
      max: "2024-03-31",
    });

    localStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  describe("Dashboards with timeseries", () => {
    const ExploreStateSubsetForBaseState: Partial<MetricsExplorerEntity> = {
      selectedTimeRange: {
        name: "rill-QTD",
        interval: V1TimeGrain.TIME_GRAIN_WEEK,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,

      visibleMeasureKeys: new Set([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]),
      allMeasuresVisible: true,
      visibleDimensionKeys: new Set([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]),
      allDimensionsVisible: true,

      leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardContextColumn: undefined,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
    };
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
      error: undefined,
      isFetching: false,
      isLoading: false,
    });

    it("Should load base dashboard state", async () => {
      renderDashboardStateLoader();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset(ExploreStateSubsetForBaseState);
      // no additional goto is called
      expect(pageMock.urlSearchHistory).toEqual([]);
    });

    it("Should load 'other source' of dashboard state", async () => {
      renderDashboardStateLoader(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForBaseState,

        selectedTimeRange: {
          name: "PT24H",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls,
      });
      const initUrlSearch = "tr=PT24H&compare_tr=rill-PP&grain=hour";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForBaseState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        "explore",
        "view=explore&tr=P7D&compare_tr=rill-PP&grain=hour&measures=bid_price&dims=domain&sort_by=bid_price",
      );
      renderDashboardStateLoader(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        selectedTimeRange: {
          name: "P7D",
          interval: V1TimeGrain.TIME_GRAIN_HOUR,
        } as DashboardTimeControls,
        showTimeComparison: true,
        selectedComparisonTimeRange: {
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls,

        visibleMeasureKeys: new Set([AD_BIDS_BID_PRICE_MEASURE]),
        allMeasuresVisible: false,
        visibleDimensionKeys: new Set([AD_BIDS_DOMAIN_DIMENSION]),
        allDimensionsVisible: false,

        leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardContextColumn: undefined,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      });
      const initUrlSearch =
        "tr=P7D&compare_tr=rill-PP&grain=hour&measures=bid_price&dims=domain&sort_by=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForBaseState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });
  });

  describe("Dashboards without timeseries", () => {
    const ExploreStateSubsetForBaseState: Partial<MetricsExplorerEntity> = {
      selectedTimeRange: undefined,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,

      visibleMeasureKeys: new Set([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]),
      allMeasuresVisible: true,
      visibleDimensionKeys: new Set([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]),
      allDimensionsVisible: true,

      leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardContextColumn: undefined,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
    };

    beforeEach(() => {
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_INIT,
        AD_BIDS_EXPLORE_INIT,
      );
    });

    it("Should load base dashboard state", async () => {
      renderDashboardStateLoader();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset(ExploreStateSubsetForBaseState);
      // no additional goto is called
      expect(pageMock.urlSearchHistory).toEqual([]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        "explore",
        "view=explore&measures=bid_price&dims=domain&sort_by=bid_price",
      );
      renderDashboardStateLoader();

      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      assertExploreStateSubset({
        selectedTimeRange: undefined,
        showTimeComparison: false,
        selectedComparisonTimeRange: undefined,

        visibleMeasureKeys: new Set([AD_BIDS_BID_PRICE_MEASURE]),
        allMeasuresVisible: false,
        visibleDimensionKeys: new Set([AD_BIDS_DOMAIN_DIMENSION]),
        allDimensionsVisible: false,

        leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardContextColumn: undefined,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      });
      const initUrlSearch = "measures=bid_price&dims=domain&sort_by=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForBaseState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });
  });
});

// This needs to be there each file because of how hoisting works with vitest.
// TODO: find if there is a way to share code.
function renderDashboardStateLoader(
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

    visibleMeasureKeys: curExploreState.visibleMeasureKeys,
    allMeasuresVisible: curExploreState.allMeasuresVisible,

    visibleDimensionKeys: curExploreState.visibleDimensionKeys,
    allDimensionsVisible: curExploreState.allDimensionsVisible,

    leaderboardMeasureName: curExploreState.leaderboardMeasureName,
    leaderboardContextColumn: curExploreState.leaderboardContextColumn,
    sortDirection: curExploreState.sortDirection,
  };
  expect(curExploreStateSubset).toEqual(exploreStateSubset);
}
