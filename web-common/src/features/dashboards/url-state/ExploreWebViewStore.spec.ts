import { QueryClient } from "@rilldata/svelte-query";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { initStateManagers } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import ExploreStateTestComponent from "@rilldata/web-common/features/dashboards/url-state/ExploreStateTestComponent.svelte";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
} from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import {
  type V1ExplorePreset,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { Page } from "@sveltejs/kit";
import { render } from "@testing-library/svelte";
import { get, type Readable, readable } from "svelte/store";
import {
  describe,
  it,
  expect,
  beforeAll,
  beforeEach,
  vi,
  afterAll,
} from "vitest";

const pageMock: Readable<Page> = vi.hoisted(() => ({}) as any);

vi.mock("$app/stores", () => {
  return {
    page: pageMock,
  };
});

const TestCases: {
  title: string;
  overviewMutations: TestDashboardMutation[];
  view: string;
  additionalPresetForView?: V1ExplorePreset;
  mutationsInView: TestDashboardMutation[];
  expectedOverviewUrl: string;
  expectedViewUrl: string;
}[] = [
  {
    title:
      "Should retain params of overview when moved to and from time dimension details",
    overviewMutations: [
      AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
      AD_BIDS_SORT_ASC_BY_BID_PRICE,
    ],
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
    additionalPresetForView: {
      timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
    },
    mutationsInView: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
    expectedOverviewUrl:
      "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_by=bid_price&sort_dir=ASC",
    expectedViewUrl:
      "/explore/AdBids_explore?view=ttd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
  },
  {
    title:
      "Should retain params of dimension table when moved to and from time dimension details",
    overviewMutations: [
      AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
      AD_BIDS_SORT_ASC_BY_BID_PRICE,
      AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
    ],
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
    additionalPresetForView: {
      timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
    },
    mutationsInView: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
    expectedOverviewUrl:
      "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=bid_price&sort_dir=ASC",
    expectedViewUrl:
      "/explore/AdBids_explore?view=ttd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
  },
];

describe("ExploreWebViewStore", () => {
  runtime.set({
    host: "http://localhost",
    instanceId: "default",
  });
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: "always",
      },
    },
  });
  let unmountTestComponent: () => void;

  beforeAll(async () => {
    const { subscribe } = readable({
      url: new URL("http://localhost/explore/" + AD_BIDS_EXPLORE_NAME),
    });
    pageMock.subscribe = subscribe as any;
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
    );
    const validSpecQuery = useExploreValidSpec(
      "default",
      AD_BIDS_EXPLORE_NAME,
      {
        queryClient,
      },
    );
    const { unmount } = render(ExploreStateTestComponent, {
      validSpecQuery,
    });
    unmountTestComponent = unmount;
    await waitUntil(() => !get(validSpecQuery).isLoading);
  });

  afterAll(() => {
    unmountTestComponent?.();
  });

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
  });

  for (const {
    title,
    overviewMutations,
    view,
    additionalPresetForView,
    mutationsInView,
    expectedOverviewUrl,
    expectedViewUrl,
  } of TestCases) {
    it(title, () => {
      metricsExplorerStore.init(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      );
      const stateManagers = createStateManagers({
        queryClient,
        metricsViewName: AD_BIDS_NAME,
        exploreName: AD_BIDS_EXPLORE_NAME,
      });
      const defaultExplorePreset = getBasePreset(
        AD_BIDS_EXPLORE_INIT,
        {
          timeZone: "UTC",
        },
        AD_BIDS_TIME_RANGE_SUMMARY,
      );

      // apply mutations to main view to setup the initial state
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, overviewMutations);
      const stateOnOverview = getCleanMetricsExploreForAssertion();

      // simulate going to the view's url
      applyURLToExploreState(
        new URL(
          "http://localhost" +
            stateManagers.webViewStore.getUrlForView(
              view as any,
              get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
              AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
              AD_BIDS_EXPLORE_INIT,
              defaultExplorePreset,
              additionalPresetForView,
            ),
        ),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      // apply any mutations in the view
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutationsInView);

      const backToOverviewUrl = stateManagers.webViewStore.getUrlForView(
        V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
        get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
        additionalPresetForView,
      );
      expect(backToOverviewUrl).toEqual(expectedOverviewUrl);
      applyURLToExploreState(
        new URL("http://localhost" + backToOverviewUrl),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );

      const backToViewUrl = stateManagers.webViewStore.getUrlForView(
        view as any,
        get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
        additionalPresetForView,
      );
      expect(backToViewUrl).toEqual(expectedViewUrl);
      const currentState = getCleanMetricsExploreForAssertion();
      expect(stateOnOverview).toEqual(currentState);
    });
  }
});
