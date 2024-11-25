import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_DIMENSION_TABLE_PRESET,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PIVOT_PRESET,
  AD_BIDS_PRESET,
  AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  AD_BIDS_CLOSE_DIMENSION_TABLE,
  AD_BIDS_CLOSE_TDD,
  AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_KATHMANDU_TIMEZONE,
  AD_BIDS_SET_LA_TIMEZONE,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  AD_BIDS_TOGGLE_PIVOT,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { describe, it, expect, beforeAll, beforeEach } from "vitest";

const AllTimeRangeKeys: (keyof MetricsExplorerEntity)[] = [
  "selectedTimeRange",
  "selectedComparisonTimeRange",
  "selectedTimezone",
];
const OverviewTestCases: {
  title: string;
  mutations: TestDashboardMutation[];
  keys: (keyof MetricsExplorerEntity)[];
  preset?: V1ExplorePreset;
  expectedUrl: string;
}[] = [
  {
    title: "Time range without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    keys: AllTimeRangeKeys,
    expectedUrl: "http://localhost/?tr=P4W&tg=week&tz=Asia%2FKathmandu",
  },
  {
    title: "Time range with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    keys: AllTimeRangeKeys,
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Time range with preset and state not matching preset",
    mutations: [AD_BIDS_SET_P4W_TIME_RANGE_FILTER, AD_BIDS_SET_LA_TIMEZONE],
    keys: AllTimeRangeKeys,
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&tg=week&tz=America%2FLos_Angeles",
  },

  {
    title: "Time range comparison without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    keys: AllTimeRangeKeys,
    expectedUrl: "http://localhost/?tr=P4W&tg=week&ctr=rill-PWC",
  },
  {
    title: "Time range comparison with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
    ],
    keys: AllTimeRangeKeys,
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Time range comparison with preset and state not matching preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    keys: AllTimeRangeKeys,
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&tg=week&ctr=rill-PWC",
  },
  {
    title: "Time range comparison enable and disable",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
    ],
    keys: AllTimeRangeKeys,
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?tr=P4W&tg=week",
  },

  {
    title:
      "Measures/dimensions visibility with no preset and partially visible measures/dimensions in state",
    mutations: [
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    keys: [],
    expectedUrl: "http://localhost/?o.m=impressions&o.d=publisher",
  },
  {
    title:
      "Measures/dimensions visibility with no preset and all measures/dimensions visible in state",
    mutations: [
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      // re-toggle to show
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    keys: [],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and partially visible measures/dimensions in state matching preset",
    mutations: [
      // initially hidden due to preset, show them now.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
      // hide them back.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    keys: [],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and all measures/dimensions visible in state not matching preset",
    mutations: [
      // initially hidden due to preset, show them now.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    keys: [],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?o.m=*&o.d=*",
  },

  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state different than default",
    mutations: [AD_BIDS_SORT_ASC_BY_BID_PRICE],
    keys: [],
    expectedUrl: "http://localhost/?o.sb=bid_price&o.sd=ASC",
  },
  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state same as default",
    mutations: [AD_BIDS_SORT_DESC_BY_IMPRESSIONS],
    keys: [],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state same as preset",
    mutations: [AD_BIDS_SORT_ASC_BY_BID_PRICE],
    keys: [],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state different than preset",
    mutations: [AD_BIDS_SORT_DESC_BY_IMPRESSIONS],
    keys: [],
    preset: AD_BIDS_PRESET,
    expectedUrl: "http://localhost/?o.sb=impressions&o.sd=DESC",
  },

  {
    title: "Dimension table with no preset and dimension table active in state",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    keys: [],
    expectedUrl: "http://localhost/?o.ed=publisher",
  },
  {
    title: "Dimension table with no preset and open and close dimension table",
    mutations: [
      AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
      AD_BIDS_CLOSE_DIMENSION_TABLE,
    ],
    keys: [],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state same as preset",
    mutations: [AD_BIDS_OPEN_DOM_DIMENSION_TABLE],
    keys: [],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state different than preset",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    keys: [],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/?o.ed=publisher",
  },
  {
    title:
      "Dimension table with preset and with no dimension table in state different than preset",
    mutations: [AD_BIDS_CLOSE_DIMENSION_TABLE],
    keys: [],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedUrl: "http://localhost/?o.ed=",
  },

  {
    title:
      "Time dimensional details with no preset and has time dimensional details in state",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
    keys: [],
    expectedUrl:
      "http://localhost/?vw=time_dimension&tdd.m=impressions&tdd.ct=stacked_bar",
  },
  {
    title: "Time dimensional details with no preset, open and close TDD",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_CLOSE_TDD],
    keys: [],
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state same as presets",
    mutations: [AD_BIDS_OPEN_IMP_TDD],
    keys: [],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state different than presets",
    mutations: [AD_BIDS_CLOSE_TDD],
    keys: [],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedUrl: "http://localhost/?vw=overview&tdd.m=",
  },

  {
    title: "Pivot with no preset and has pivot in state",
    mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
    keys: [],
    expectedUrl:
      "http://localhost/?vw=pivot&p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
  },
  {
    title: "Pivot with no preset, open and close pivot",
    mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS, AD_BIDS_TOGGLE_PIVOT],
    keys: [],
    expectedUrl:
      "http://localhost/?p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
  },
  {
    title: "Pivot with preset and has pivot in state same as preset",
    mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
    keys: [],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl: "http://localhost/",
  },
  {
    title: "Pivot with preset and pivot in state different as preset",
    mutations: [AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT],
    keys: [],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl: "http://localhost/?p.r=domain%2Ctime.day&p.c=impressions",
  },
  {
    title: "Pivot with preset and no pivot in state different as preset",
    mutations: [AD_BIDS_TOGGLE_PIVOT],
    keys: [],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedUrl: "http://localhost/?vw=overview",
  },
];

describe("Human readable URL state variations", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
  });

  describe("Should update url state and restore default state on empty params", () => {
    for (const { title, mutations, preset, expectedUrl } of OverviewTestCases) {
      it(title, () => {
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
        metricsExplorerStore.init(
          AD_BIDS_EXPLORE_NAME,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );
        const initState = getCleanMetricsExploreForAssertion();
        const basePreset = getBasePreset(
          explore,
          {
            timeZone: "UTC",
          },
          AD_BIDS_TIME_RANGE_SUMMARY,
        );

        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        // load url params with updated metrics state
        const url = new URL("http://localhost");
        mergeSearchParams(
          convertMetricsEntityToURLSearchParams(
            get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
            explore,
            basePreset,
          ),
          url.searchParams,
        );
        expect(url.toString()).to.eq(expectedUrl);

        // load empty url into metrics
        const defaultUrl = new URL("http://localhost");
        const { partialExploreState: partialExploreStateDefaultUrl } =
          convertURLToMetricsExplore(
            defaultUrl.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            basePreset,
          );
        metricsExplorerStore.mergePartialExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          partialExploreStateDefaultUrl,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        );
        const currentState = getCleanMetricsExploreForAssertion();
        // current state should match the initial state
        expect(currentState).toEqual(initState);
      });
    }
  });
});

export function getCleanMetricsExploreForAssertion() {
  const cleanedState = deepClone(
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
  ) as Partial<MetricsExplorerEntity>;
  // these are not cloned
  cleanedState.visibleMeasureKeys = new Set(cleanedState.visibleMeasureKeys);
  cleanedState.visibleDimensionKeys = new Set(
    cleanedState.visibleDimensionKeys,
  );

  delete cleanedState.name;
  delete cleanedState.proto;
  delete cleanedState.dimensionFilterExcludeMode;
  delete cleanedState.temporaryFilterName;
  delete cleanedState.contextColumnWidths;
  if (cleanedState.selectedTimeRange) {
    cleanedState.selectedTimeRange = {
      name: cleanedState.selectedTimeRange?.name ?? "inf",
      interval: cleanedState.selectedTimeRange?.interval,
    } as DashboardTimeControls;
  }
  delete cleanedState.lastDefinedScrubRange;

  // TODO
  delete cleanedState.selectedScrubRange;
  delete cleanedState.leaderboardContextColumn;
  delete cleanedState.dashboardSortType;

  return cleanedState;
}
