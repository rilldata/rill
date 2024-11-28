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
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
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
import { ExploreWebViewNonPivot } from "@rilldata/web-common/features/dashboards/url-state/ExploreWebViewStore";
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

type TestView = {
  view: string;
  previousView?: string;
  additionalPresetForView?: V1ExplorePreset;
  mutations: TestDashboardMutation[];
  expectedUrl: string;
};
const TestCases: {
  title: string;
  initView: TestView;
  view: TestView;
}[] = [
  {
    title: "overview <=> TTD",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
      mutations: [],
      expectedUrl:
        "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_by=bid_price&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalPresetForView: {
        timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      },
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "/explore/AdBids_explore?view=ttd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },
  {
    title: "dimension table <=> TTD",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedUrl:
        "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=bid_price&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalPresetForView: {
        timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      },
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "/explore/AdBids_explore?view=ttd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
  },

  {
    title: "overview <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
      mutations: [],
      expectedUrl:
        "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&sort_by=bid_price&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      previousView: ExploreWebViewNonPivot,
      mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
      expectedUrl:
        "/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions",
    },
  },
  {
    title: "dimension table <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
      mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
      expectedUrl:
        "/explore/AdBids_explore?tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measures=impressions&dims=publisher&expand_dim=publisher&sort_by=bid_price&sort_dir=ASC",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      previousView: ExploreWebViewNonPivot,
      mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
      expectedUrl:
        "/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions",
    },
  },
  {
    title: "TTD <=> Pivot",
    initView: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      additionalPresetForView: {
        timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      },
      mutations: [AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
      expectedUrl:
        "/explore/AdBids_explore?view=ttd&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&measure=impressions&chart_type=stacked_bar",
    },
    view: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      previousView: ExploreWebViewNonPivot,
      mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
      expectedUrl:
        "/explore/AdBids_explore?view=pivot&tr=P7D&compare_tr=rill-PP&grain=day&f=publisher+IN+%28%27Google%27%29&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions",
    },
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

  for (const { title, initView, view } of TestCases) {
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
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, [
        AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
        AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
        AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
        AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
        AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
        AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
        AD_BIDS_SORT_ASC_BY_BID_PRICE,
      ]);

      // simulate going to the init view's url
      applyURLToExploreState(
        new URL(
          "http://localhost" +
            stateManagers.webViewStore.getUrlForView(
              initView.view as any,
              get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
              AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
              AD_BIDS_EXPLORE_INIT,
              defaultExplorePreset,
              initView.additionalPresetForView,
            ),
        ),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      // apply any mutations in the init view
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, initView.mutations);
      const initState = getCleanMetricsExploreForAssertion();

      // simulate going to the view's url
      applyURLToExploreState(
        new URL(
          "http://localhost" +
            stateManagers.webViewStore.getUrlForView(
              view.view as any,
              get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
              AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
              AD_BIDS_EXPLORE_INIT,
              defaultExplorePreset,
              view.additionalPresetForView,
            ),
        ),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      // apply any mutations in the view
      applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, view.mutations);

      const backToInitUrl = stateManagers.webViewStore.getUrlForView(
        (view.previousView ?? initView.view) as any,
        get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      expect(backToInitUrl).toEqual(initView.expectedUrl);
      applyURLToExploreState(
        new URL("http://localhost" + backToInitUrl),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );

      const backToViewUrl = stateManagers.webViewStore.getUrlForView(
        view.view as any,
        get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
      );
      expect(backToViewUrl).toEqual(view.expectedUrl);
      const currentState = getCleanMetricsExploreForAssertion();
      expect(initState).toEqual(currentState);
    });
  }
});
