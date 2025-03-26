import { useDashboardFetchMocksForComponentTests } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
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
  createPageMock,
  type PageMock,
} from "@rilldata/web-common/features/dashboards/url-state/PageMock";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import DashboardStateLoaderTest from "./DashboardStateLoaderTest.svelte";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { render, waitFor, screen } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

const pageMock: PageMock = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url) => pageMock.goto(url),
  };
});
vi.mock("$app/stores", () => {
  return {
    page: pageMock,
  };
});

describe("DashboardStateLoader", () => {
  mockAnimationsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  beforeEach(() => {
    createPageMock(pageMock);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, {
      min: "2025-01-01",
      max: "2025-03-31",
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  it("Should load base dashboard state for metrics view with timeseries", async () => {
    renderDashboardStateLoader();

    await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
    const metricsView = getCleanMetricsExploreForAssertion();
    expect(metricsView.selectedTimeRange?.name).toEqual("rill-QTD");
    expect([...metricsView.visibleMeasureKeys!]).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_BID_PRICE_MEASURE,
    ]);
    expect([...metricsView.visibleDimensionKeys!]).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(metricsView.leaderboardMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
  });

  it("Should load base dashboard state for metrics view without timeseries", async () => {
    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT,
      AD_BIDS_EXPLORE_INIT,
    );
    renderDashboardStateLoader();

    await waitFor(() => expect(screen.getByText("Dashboard loaded!")));
    const metricsView = getCleanMetricsExploreForAssertion();
    expect(metricsView.selectedTimeRange?.name).toBeUndefined();
    expect([...metricsView.visibleMeasureKeys!]).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_BID_PRICE_MEASURE,
    ]);
    expect([...metricsView.visibleDimensionKeys!]).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(metricsView.leaderboardMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
  });
});

function renderDashboardStateLoader() {
  const renderResults = render(DashboardStateLoaderTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map([["$$_queryClient", queryClient]]),
  });

  return { queryClient, renderResults };
}
