import { type CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { useDashboardFetchMocksForComponentTests } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { setExploreStateForWebView } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { setMostRecentExploreStateInLocalStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import DashboardStateManagerTest from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/DashboardStateManagerTest.svelte";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { resolveTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  type V1ExploreSpec,
  V1Operation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { readable } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
  };
});

// Stands in for the network call that resolves time ranges, so a test can hold it open.
const hoistedTimeRanges = vi.hoisted(() => ({
  pending: Promise.resolve(),
}));
vi.mock(
  "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges",
  async (importOriginal) => ({
    ...(await importOriginal<object>()),
    resolveTimeRanges: vi.fn(
      async (
        _client: unknown,
        _exploreSpec: unknown,
        timeRanges: unknown[],
      ) => {
        await hoistedTimeRanges.pending;
        return timeRanges;
      },
    ),
  }),
);
const VitestEnv = String(import.meta.env.VITEST);

// Contains basic tests for verifying order of selection.
// In-depth tests for storing and retrieving state are separate.
describe("DashboardStateManager", () => {
  mockAnimationsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForComponentTests;

  beforeEach(() => {
    pageMock = new PageMockForComponentTests(hoistedPage);

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
    const ExploreStateSubsetForRillDefaultState: Partial<ExploreState> = {
      selectedTimeRange: {
        name: "rill-QTD",
        interval: V1TimeGrain.TIME_GRAIN_WEEK,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,

      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: true,

      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
      dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
    };
    const ExploreStateSubsetForYAMLState: Partial<ExploreState> = {
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

      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    };
    const PageURLForRillDefaultState =
      "tr=P7D&tz=Asia%2FKathmandu&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";
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

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,
        ...ExploreStateSubsetForYAMLState,
      });

      pageMock.assertSearchParams(PageURLForRillDefaultState);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        PageURLForRillDefaultState,
        "",
      ]);
    });

    it("Should load 'other source' of dashboard state", async () => {
      renderDashboardStateManager(BookmarkSourceQueryResult);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,
        ...ExploreStateSubsetForYAMLState,

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
        "tr=PT24H&tz=Asia%2FKathmandu&compare_tr=rill-PP&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should drop pivot fields repeated in the 'other source' of dashboard state", async () => {
      // A legacy home bookmark loads through proto state, not the url parser,
      // so the final correction has to dedupe what the merge hands it.
      const publisher: PivotChipData = {
        id: AD_BIDS_PUBLISHER_DIMENSION,
        title: AD_BIDS_PUBLISHER_DIMENSION,
        type: PivotChipType.Dimension,
      };
      const domain: PivotChipData = {
        id: AD_BIDS_DOMAIN_DIMENSION,
        title: AD_BIDS_DOMAIN_DIMENSION,
        type: PivotChipType.Dimension,
      };
      const impressions: PivotChipData = {
        id: AD_BIDS_IMPRESSIONS_MEASURE,
        title: AD_BIDS_IMPRESSIONS_MEASURE,
        type: PivotChipType.Measure,
      };
      renderDashboardStateManager(
        readable({
          data: {
            activePage: DashboardState_ActivePage.PIVOT,
            pivot: {
              rows: [publisher, publisher],
              columns: [publisher, domain, impressions, impressions],
            } as PivotState,
          },
          error: null,
          isFetching: false,
          isLoading: false,
        }),
      );
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      const { pivot } = getCleanMetricsExploreForAssertion();
      expect(pivot?.rows).toEqual([publisher]);
      expect(pivot?.columns).toEqual([domain, impressions]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreStateInLocalStorage(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
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
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
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
        "tr=PT24H&tz=Asia%2FKathmandu&compare_tr=rill-PP&measures=bid_price&dims=domain&sort_by=bid_price&sort_dir=ASC&leaderboard_measures=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should load from session dashboard state", async () => {
      setExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        ExploreUrlWebView.Explore,
        "view=explore&tr=P14D&compare_tr=rill-PW&grain=day&measures=bid_price&dims=domain&sort_by=bid_price&sort_type=delta_abs&sort_dir=DESC&leaderboard_measures=bid_price",
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
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
        dashboardSortType: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      });
      const initUrlSearch =
        "tr=P14D&tz=Asia%2FKathmandu&compare_tr=rill-PW&grain=day&measures=bid_price&dims=domain&sort_by=bid_price&sort_type=delta_abs&leaderboard_measures=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should not load local storage or session storage when disabled", async () => {
      setMostRecentExploreStateInLocalStorage(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      setExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        ExploreUrlWebView.Explore,
        "view=explore&tr=P14D&compare_tr=rill-PW&grain=day&measures=bid_price&dims=domain&sort_by=bid_price&sort_type=delta_abs&sort_dir=DESC&leaderboard_measures=bid_price",
      );
      renderDashboardStateManager(undefined, true, true);
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,
        ...ExploreStateSubsetForYAMLState,
      });
    });

    describe("Url changes while a time range resolves", () => {
      let releaseTimeRanges!: () => void;

      beforeEach(() => {
        // DashboardStateSync skips resolving time ranges under vitest.
        vi.stubEnv("VITEST", "");
        vi.mocked(resolveTimeRanges).mockClear();
        hoistedTimeRanges.pending = new Promise(
          (resolve) => (releaseTimeRanges = resolve),
        );
      });

      afterEach(() => {
        vi.stubEnv("VITEST", VitestEnv);
        hoistedTimeRanges.pending = Promise.resolve();
      });

      it("Should keep a url change made while the dashboard initializes", async () => {
        renderDashboardStateManager();
        await waitFor(() => expect(resolveTimeRanges).toHaveBeenCalledOnce());

        // Like an embed's setValidState that lands before the dashboard has loaded.
        pageMock.gotoSearch("tr=PT24H");
        releaseTimeRanges();

        await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
        expect(
          getCleanMetricsExploreForAssertion().selectedTimeRange?.name,
        ).toEqual("PT24H");
        pageMock.assertSearchParams("tr=PT24H");
      });

      it("Should keep a url change made while an earlier one is applied", async () => {
        await startUrlChangeWithPendingTimeRange();
        pageMock.gotoSearch("tr=P14D");
        releaseTimeRanges();

        await waitFor(() =>
          expect(
            getCleanMetricsExploreForAssertion().selectedTimeRange?.name,
          ).toEqual("P14D"),
        );
        pageMock.assertSearchParams("tr=P14D");
      });

      it("Should apply going back while an earlier url change is applied like any going back", async () => {
        await startUrlChangeWithPendingTimeRange();
        pageMock.popState("");
        releaseTimeRanges();

        // Going back to an empty url gives the rill defaults, not the session state.
        await waitFor(() =>
          assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
        );
        pageMock.assertSearchParams("");
      });

      it("Should drop a url change when another explore opens while it is applied", async () => {
        await startUrlChangeWithPendingTimeRange();
        hoistedPage.goto(
          new URL("http://localhost/explore/other_explore?dims=other_dim"),
        );
        releaseTimeRanges();
        await asyncWait(10);

        pageMock.assertSearchParams("dims=other_dim");
        expect(
          getCleanMetricsExploreForAssertion().selectedTimeRange?.name,
        ).toEqual("P7D");
      });

      it("Should drop a url change when the dashboard is torn down while it is applied", async () => {
        const renderResults = await startUrlChangeWithPendingTimeRange();
        renderResults.unmount();
        releaseTimeRanges();
        await asyncWait(10);

        expect(
          getCleanMetricsExploreForAssertion().selectedTimeRange?.name,
        ).toEqual("P7D");
      });

      // Loads the dashboard, then starts a url change whose time range stays unresolved.
      async function startUrlChangeWithPendingTimeRange() {
        releaseTimeRanges();
        const { renderResults } = renderDashboardStateManager();
        await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

        hoistedTimeRanges.pending = new Promise(
          (resolve) => (releaseTimeRanges = resolve),
        );
        pageMock.gotoSearch("tr=PT24H");
        await waitFor(() => expect(resolveTimeRanges).toHaveBeenCalledTimes(2));
        return renderResults;
      }
    });
  });

  describe("Dashboards with an explore spec naming fields the metrics view dropped", () => {
    // With staged changes the runtime keeps serving the explore's last valid spec
    // after the metrics view dropped a field, so the explore lists fields that cannot be queried.
    const StaleExploreSpec: V1ExploreSpec = {
      ...AD_BIDS_EXPLORE_INIT,
      measures: [
        AD_BIDS_IMPRESSIONS_MEASURE,
        "stale_measure",
        AD_BIDS_BID_PRICE_MEASURE,
      ],
      dimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        "stale_dimension",
        AD_BIDS_DOMAIN_DIMENSION,
      ],
    };
    const ExploreStateSubsetForRillDefaultState: Partial<ExploreState> = {
      selectedTimeRange: {
        name: "rill-QTD",
        interval: V1TimeGrain.TIME_GRAIN_WEEK,
      } as DashboardTimeControls,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,

      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: true,

      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
      dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
    };

    it("Should drop the stale fields from the rill default state", async () => {
      mocks.mockMetricsExplore(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_INIT_WITH_TIME,
        StaleExploreSpec,
      );
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset(ExploreStateSubsetForRillDefaultState);
      pageMock.assertSearchParams("");
    });

    it("Should drop the stale fields from the yaml defaults", async () => {
      mocks.mockMetricsExplore(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_INIT_WITH_TIME,
        {
          ...StaleExploreSpec,
          defaultPreset: {
            measures: ["stale_measure", AD_BIDS_BID_PRICE_MEASURE],
            dimensions: ["stale_dimension", AD_BIDS_DOMAIN_DIMENSION],
            exploreSortBy: "stale_measure",
            exploreLeaderboardMeasures: ["stale_measure"],
            comparisonMode:
              V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
          },
        },
      );
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
      });
      pageMock.assertSearchParams(
        "measures=bid_price&dims=domain&sort_by=bid_price&leaderboard_measures=bid_price",
      );
    });
  });

  describe("Dashboards with a default filter", () => {
    const DefaultFilter = createAndExpression([
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Contoso"]),
    ]);
    const OtherFilter = createAndExpression([
      createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["fabrikam.com"]),
    ]);

    beforeEach(() => {
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
        ...AD_BIDS_EXPLORE_INIT,
        requiredFilters: [AD_BIDS_PUBLISHER_DIMENSION],
        defaultPreset: {
          // What the runtime parses from `defaults.filters: publisher = 'Contoso'`.
          where: createBinaryExpression(
            AD_BIDS_PUBLISHER_DIMENSION,
            V1Operation.OPERATION_EQ,
            "Contoso",
          ),
        },
      });
    });

    it("Should apply the default filter on first load and write it to the url", async () => {
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      expect(getCleanMetricsExploreForAssertion().whereFilter).toEqual(
        DefaultFilter,
      );
      pageMock.assertSearchParams("f=publisher IN ('Contoso')");
    });

    it("Should let a filter in the url win over the default filter", async () => {
      pageMock.gotoSearch("f=domain IN ('fabrikam.com')");
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      expect(getCleanMetricsExploreForAssertion().whereFilter).toEqual(
        OtherFilter,
      );
      pageMock.assertSearchParams("f=domain IN ('fabrikam.com')");
    });

    it("Should let the filter of a bookmark win over the default filter", async () => {
      renderDashboardStateManager(
        readable({
          data: { whereFilter: OtherFilter, dimensionsWithInlistFilter: [] },
          error: null,
          isFetching: false,
          isLoading: false,
        }),
      );
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      expect(getCleanMetricsExploreForAssertion().whereFilter).toEqual(
        OtherFilter,
      );
      pageMock.assertSearchParams("f=domain IN ('fabrikam.com')");
    });
  });

  describe("Dashboards without timeseries", () => {
    const ExploreStateSubsetForRillDefaultState: Partial<ExploreState> = {
      selectedTimeRange: undefined,
      showTimeComparison: false,
      selectedComparisonTimeRange: undefined,

      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: true,
      visibleDimensions: [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ],
      allDimensionsVisible: true,

      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
    };
    const ExploreStateSubsetForYAMLState: Partial<ExploreState> = {
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      allMeasuresVisible: false,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      allDimensionsVisible: false,

      sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
      dashboardSortType: DashboardState_LeaderboardSortType.PERCENT,
    };
    const PageURLForRillDefaultState =
      "measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";

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

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,
        ...ExploreStateSubsetForYAMLState,
      });

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([
        PageURLForRillDefaultState,
        "",
      ]);
    });

    it("Should load most recent dashboard state", async () => {
      setMostRecentExploreStateInLocalStorage(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
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
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      const initUrlSearch =
        "measures=bid_price&dims=domain&sort_by=bid_price&sort_dir=ASC&leaderboard_measures=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should validate most recent dashboard state and correct invalid fields", async () => {
      setMostRecentExploreStateInLocalStorage(AD_BIDS_EXPLORE_NAME, undefined, {
        visibleMeasures: [AD_BIDS_PUBLISHER_COUNT_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_COUNTRY_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_PUBLISHER_COUNT_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_PUBLISHER_COUNT_MEASURE],
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      renderDashboardStateManager();
      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

      assertExploreStateSubset({
        ...ExploreStateSubsetForRillDefaultState,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
        sortDirection: DashboardState_LeaderboardSortDirection.ASCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
      });
      const initUrlSearch = "measures=impressions&dims=publisher&sort_dir=ASC";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });

    it("Should load from session dashboard state", async () => {
      setExploreStateForWebView(
        AD_BIDS_EXPLORE_NAME,
        undefined,
        ExploreUrlWebView.Explore,
        "view=explore&measures=bid_price&dims=domain&sort_by=bid_price&sort_type=delta_abs&sort_dir=DESC&leaderboard_measures=bid_price",
      );
      renderDashboardStateManager();

      await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
      assertExploreStateSubset({
        selectedComparisonTimeRange: undefined,
        selectedTimeRange: undefined,
        showTimeComparison: false,

        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
        allDimensionsVisible: false,

        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        leaderboardMeasureNames: [AD_BIDS_BID_PRICE_MEASURE],
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
        dashboardSortType: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
      });
      const initUrlSearch =
        "measures=bid_price&dims=domain&sort_by=bid_price&sort_type=delta_abs&leaderboard_measures=bid_price";
      pageMock.assertSearchParams(initUrlSearch);

      pageMock.popState("");
      await waitFor(() =>
        assertExploreStateSubset(ExploreStateSubsetForRillDefaultState),
      );
      // only 2 urls should in history
      expect(pageMock.urlSearchHistory).toEqual([initUrlSearch, ""]);
    });
  });
});

// This needs to be there each file because of how hoisting works with vitest.
// TODO: find if there is a way to share code.
function renderDashboardStateManager(
  bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<ExploreState> | undefined>
    | undefined = undefined,
  disableMostRecentDashboardState: boolean = false,
  disableInitSessionDashboardState: boolean = false,
) {
  const renderResults = render(DashboardStateManagerTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
      bookmarkOrTokenExploreState,
      disableMostRecentDashboardState,
      disableInitSessionDashboardState,
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map<string | symbol, unknown>([
      ["$$_queryClient", queryClient],
      [
        RUNTIME_CONTEXT_KEY,
        new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
      ],
    ]),
  });

  return { queryClient, renderResults };
}

function assertExploreStateSubset(exploreStateSubset: Partial<ExploreState>) {
  const curExploreState = getCleanMetricsExploreForAssertion();
  const curExploreStateSubset: Partial<ExploreState> = {
    selectedTimeRange: curExploreState.selectedTimeRange,
    showTimeComparison: curExploreState.showTimeComparison,
    selectedComparisonTimeRange: curExploreState.selectedComparisonTimeRange,

    visibleMeasures: curExploreState.visibleMeasures,
    allMeasuresVisible: curExploreState.allMeasuresVisible,

    visibleDimensions: curExploreState.visibleDimensions,
    allDimensionsVisible: curExploreState.allDimensionsVisible,

    leaderboardSortByMeasureName: curExploreState.leaderboardSortByMeasureName,
    leaderboardMeasureNames: curExploreState.leaderboardMeasureNames,
    dashboardSortType: curExploreState.dashboardSortType,
    sortDirection: curExploreState.sortDirection,
  };
  expect(curExploreStateSubset).toEqual(exploreStateSubset);
}
