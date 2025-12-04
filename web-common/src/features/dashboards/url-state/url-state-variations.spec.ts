import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_DIMENSION_TABLE_PRESET,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PIVOT_PRESET,
  AD_BIDS_PRESET,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_APPLY_DOMAIN_CONTAINS_FILTER,
  AD_BIDS_APPLY_LARGE_FILTERS,
  AD_BIDS_APPLY_PUBLISHER_INLIST_FILTER,
  AD_BIDS_CLOSE_DIMENSION_TABLE,
  AD_BIDS_CLOSE_TDD,
  AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_LARGE_FILTER,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_ALL_TIME_RANGE_FILTER,
  AD_BIDS_SET_KATHMANDU_TIMEZONE,
  AD_BIDS_SET_LA_TIMEZONE,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_ASC_BY_IMPRESSIONS,
  AD_BIDS_SORT_BY_DELTA_ABS_VALUE,
  AD_BIDS_SORT_BY_PERCENT_VALUE,
  AD_BIDS_SORT_BY_VALUE,
  AD_BIDS_SORT_DESC_BY_BID_PRICE,
  AD_BIDS_SORT_DESC_BY_IMPRESSIONS,
  AD_BIDS_SORT_PIVOT_BY_IMPRESSIONS_DESC,
  AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
  AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD,
  AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
  AD_BIDS_TOGGLE_BID_PUBLISHER_DIMENSION_VISIBILITY,
  AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
  AD_BIDS_TOGGLE_PIVOT,
  AD_BIDS_FLAT_PIVOT_TABLE,
  applyMutationsToDashboard,
  type TestDashboardMutation,
  AD_BIDS_SET_PUBLISHER_COMPARE_DIMENSION,
  AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION,
  AD_BIDS_MEASURE_NAMES_BID_PRICE_AND_IMPRESSIONS,
  AD_BIDS_APPLY_IMP_COUNTRY_BETWEEN_MEASURE_FILTER,
  AD_BIDS_APPLY_IMP_COUNTRY_NOT_BETWEEN_MEASURE_FILTER,
  AD_BIDS_SET_MINUTE_TIME_GRAIN,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getCleanedUrlParamsForGoto } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ALL_TIME_RANGE_ALIAS } from "../time-controls/new-time-controls";
import { convertURLSearchParamsToExploreState } from "./convertURLSearchParamsToExploreState";

vi.stubEnv("TZ", "UTC");

const TestCases: {
  title: string;
  mutations: TestDashboardMutation[];
  preset?: V1ExplorePreset;
  expectedSearch: string;
  // This is to assert edge case when some state gets populated from timeControlStore
  extraExploreState?: Partial<ExploreState>;
  // Mainly tests that close certain views.
  // Closing view would retain some state of the old view in protobuf state
  legacyNotSupported?: boolean;
}[] = [
  {
    title: "Different filter variations",
    mutations: [
      AD_BIDS_APPLY_PUBLISHER_INLIST_FILTER,
      AD_BIDS_APPLY_DOMAIN_CONTAINS_FILTER,
      AD_BIDS_APPLY_IMP_COUNTRY_BETWEEN_MEASURE_FILTER,
    ],
    expectedSearch:
      "f=publisher+IN+LIST+%28%27Facebook%27%2C%27Google%27%29+AND+domain+LIKE+%27%25%25oo%25%25%27+AND+country+having+%28%28bid_price+GT+10+AND+bid_price+LT+20%29%29",
  },
  {
    title: "Not-between measure filter",
    mutations: [AD_BIDS_APPLY_IMP_COUNTRY_NOT_BETWEEN_MEASURE_FILTER],
    expectedSearch:
      "f=country+having+%28%28bid_price+LTE+10+OR+bid_price+GTE+20%29%29",
  },

  {
    title: "Time range without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    expectedSearch: "tr=P4W&tz=Asia%2FKathmandu&grain=week",
  },
  {
    title: "Time range with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_KATHMANDU_TIMEZONE,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PW&grain=day&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
    extraExploreState: {
      selectedComparisonTimeRange: {
        name: "rill-PP",
      } as DashboardTimeControls,
    },
  },
  {
    title: "Time range with preset and state not matching preset",
    mutations: [AD_BIDS_SET_P4W_TIME_RANGE_FILTER, AD_BIDS_SET_LA_TIMEZONE],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P4W&tz=America%2FLos_Angeles&compare_tr=rill-PP&grain=week&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
    extraExploreState: {
      selectedComparisonTimeRange: {
        name: "rill-PP",
      } as DashboardTimeControls,
    },
  },
  {
    title: "Time range with preset and ALL_TIME selected",
    mutations: [AD_BIDS_SET_ALL_TIME_RANGE_FILTER],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=inf&tz=Asia%2FKathmandu&grain=day&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
  },

  {
    title: "Time range comparison without preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    expectedSearch: "tr=P4W&compare_tr=rill-PW&grain=week",
  },
  {
    title: "Time range comparison with preset and state matching preset",
    mutations: [
      AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=day&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
  },
  {
    title: "Time range comparison with preset and state not matching preset",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P4W&tz=Asia%2FKathmandu&compare_tr=rill-PW&grain=week&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
  },
  {
    title: "Time range comparison enable and disable",
    mutations: [
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_WEEK_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P4W&tz=Asia%2FKathmandu&grain=week&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
    legacyNotSupported: true,
  },
  {
    title: "Time range comparison with non-standard time range in preset",
    mutations: [AD_BIDS_DISABLE_COMPARE_TIME_RANGE_FILTER],
    preset: {
      timeRange: "P9D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
    },
    expectedSearch: "tr=P9D&grain=day",
    legacyNotSupported: true,
  },

  {
    title: "Only time grain different than default",
    mutations: [AD_BIDS_SET_MINUTE_TIME_GRAIN],
    expectedSearch: "grain=minute",
  },

  {
    title: "Dimension comparison without preset",
    mutations: [AD_BIDS_SET_PUBLISHER_COMPARE_DIMENSION],
    expectedSearch: "compare_dim=publisher",
  },
  {
    title: "Dimension comparison with preset and matching preset",
    mutations: [AD_BIDS_SET_PUBLISHER_COMPARE_DIMENSION],
    preset: {
      comparisonDimension: AD_BIDS_PUBLISHER_DIMENSION,
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION,
    },
    expectedSearch: "compare_dim=publisher",
  },
  {
    title: "Dimension comparison with preset and not matching preset",
    mutations: [AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION],
    preset: {
      comparisonDimension: AD_BIDS_PUBLISHER_DIMENSION,
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION,
    },
    expectedSearch: "compare_dim=domain",
  },

  {
    title:
      "Measures/dimensions visibility with no preset and partially visible measures/dimensions in state",
    mutations: [
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    expectedSearch: "measures=impressions&dims=publisher",
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
    expectedSearch: "",
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
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=day&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
  },
  {
    title:
      "Measures/dimensions visibility with preset and all measures/dimensions visible in state not matching preset",
    mutations: [
      // initially hidden due to preset, show them now.
      AD_BIDS_TOGGLE_BID_PRICE_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION_VISIBILITY,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=day&sort_type=percent&sort_dir=ASC",
  },
  {
    title: "Show and hide measures/dimensions",
    mutations: [
      AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_IMPRESSIONS_MEASURE_VISIBILITY,
      AD_BIDS_TOGGLE_BID_PUBLISHER_DIMENSION_VISIBILITY,
      AD_BIDS_TOGGLE_BID_PUBLISHER_DIMENSION_VISIBILITY,
    ],
    expectedSearch:
      "measures=bid_price%2Cimpressions&dims=domain%2Cpublisher&sort_by=bid_price",
  },

  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state different than default",
    mutations: [AD_BIDS_SORT_BY_DELTA_ABS_VALUE, AD_BIDS_SORT_ASC_BY_BID_PRICE],
    expectedSearch: "sort_by=bid_price&sort_type=delta_abs&sort_dir=ASC",
  },
  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state same as default",
    mutations: [AD_BIDS_SORT_BY_VALUE, AD_BIDS_SORT_DESC_BY_IMPRESSIONS],
    expectedSearch: "",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state same as preset",
    mutations: [AD_BIDS_SORT_BY_PERCENT_VALUE, AD_BIDS_SORT_ASC_BY_IMPRESSIONS],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=day&measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state different than preset",
    mutations: [
      AD_BIDS_SORT_BY_DELTA_ABS_VALUE,
      AD_BIDS_SORT_DESC_BY_BID_PRICE,
    ],
    preset: AD_BIDS_PRESET,
    expectedSearch:
      "tr=P7D&tz=Asia%2FKathmandu&compare_tr=rill-PP&grain=day&measures=impressions&dims=publisher&sort_by=bid_price&sort_type=delta_abs",
  },
  {
    title: "Leaderboard configs with multiple measures",
    mutations: [AD_BIDS_MEASURE_NAMES_BID_PRICE_AND_IMPRESSIONS],
    expectedSearch: "leaderboard_measures=bid_price%2Cimpressions",
  },

  {
    title: "Dimension table with no preset and dimension table active in state",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    expectedSearch: "expand_dim=publisher",
  },
  {
    title: "Dimension table with no preset and open and close dimension table",
    mutations: [
      AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
      AD_BIDS_CLOSE_DIMENSION_TABLE,
    ],
    expectedSearch: "",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state same as preset",
    mutations: [AD_BIDS_OPEN_DOM_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedSearch: "expand_dim=domain",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state different than preset",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedSearch: "expand_dim=publisher",
  },
  {
    title:
      "Dimension table with preset and with no dimension table in state different than preset",
    mutations: [AD_BIDS_CLOSE_DIMENSION_TABLE],
    preset: AD_BIDS_DIMENSION_TABLE_PRESET,
    expectedSearch: "",
    legacyNotSupported: true,
  },

  {
    title:
      "Time dimensional details with no preset and has time dimensional details in state",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SWITCH_TO_STACKED_BAR_IN_TDD],
    expectedSearch: "view=tdd&measure=impressions&chart_type=stacked_bar",
  },
  {
    title: "Time dimensional details with no preset, open and close TDD",
    mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_CLOSE_TDD],
    expectedSearch: "",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state same as presets",
    mutations: [AD_BIDS_OPEN_IMP_TDD],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedSearch: "view=tdd&measure=impressions&chart_type=stacked_bar",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state different than presets",
    mutations: [AD_BIDS_CLOSE_TDD],
    preset: AD_BIDS_TIME_DIMENSION_DETAILS_PRESET,
    expectedSearch: "",
    legacyNotSupported: true,
  },

  {
    title: "Pivot with no preset and has pivot in state",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
    ],
    expectedSearch:
      "view=pivot&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC&table_mode=nest",
  },
  {
    title: "Pivot with no preset, open and close pivot",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
      AD_BIDS_TOGGLE_PIVOT,
    ],
    expectedSearch: "",
    legacyNotSupported: true,
  },
  {
    title: "Pivot with no preset, toggle pivot to flat mode",
    mutations: [AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT, AD_BIDS_FLAT_PIVOT_TABLE],
    expectedSearch:
      "view=pivot&cols=domain%2Ctime.day%2Cimpressions&sort_by=&table_mode=flat",
    legacyNotSupported: true,
  },
  {
    title: "Pivot with preset and has pivot in state same as preset",
    mutations: [
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
      AD_BIDS_SORT_PIVOT_BY_TIME_DAY_ASC,
    ],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedSearch:
      "view=pivot&rows=publisher%2Ctime.hour&cols=domain%2Ctime.day%2Cimpressions&sort_by=time.day&sort_dir=ASC&table_mode=nest",
  },
  {
    title: "Pivot with preset and pivot in state different as preset",
    mutations: [
      AD_BIDS_OPEN_DOMAIN_BID_PRICE_PIVOT,
      AD_BIDS_SORT_PIVOT_BY_IMPRESSIONS_DESC,
    ],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedSearch:
      "view=pivot&rows=domain%2Ctime.day&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
  },
  {
    title: "Pivot with preset and no pivot in state different as preset",
    mutations: [AD_BIDS_TOGGLE_PIVOT],
    preset: AD_BIDS_PIVOT_PRESET,
    expectedSearch: "",
    legacyNotSupported: true,
  },
];

describe("Human readable URL state variations", () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  describe("Should update url state and restore default state on empty params", () => {
    for (const { title, mutations, preset, expectedSearch } of TestCases) {
      it(title, async () => {
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
          timeZones: ["UTC", "Asia/Kathmandu"],
        };
        metricsExplorerStore.init(
          AD_BIDS_EXPLORE_NAME,
          getInitExploreStateForTest(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY,
          ),
        );
        const initState = getCleanMetricsExploreForAssertion();
        const defaultExploreUrlSearch = getRillDefaultExploreUrlParams(
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        );
        const defaultExplorePreset = getDefaultExplorePreset(
          explore,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
          AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        );

        await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        // load url params with updated metrics state
        const updateUrlParams = getCleanedUrlParamsForGoto(
          explore,
          get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          getTimeControlState(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
            get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          ),
          defaultExploreUrlSearch,
        );
        expect(updateUrlParams.toString()).to.eq(expectedSearch);

        // load empty url into metrics
        const defaultUrl = new URL("http://localhost");
        const errors = applyURLToExploreState(
          defaultUrl,
          explore,
          defaultExplorePreset,
        );
        expect(errors.length).toEqual(0);
        const currentState = getCleanMetricsExploreForAssertion();
        // current state should match the initial state
        expect(currentState).toEqual(initState);
      });
    }
  });

  describe("Should set correct state for legacy protobuf state and restore default state on empty params", () => {
    for (const {
      title,
      mutations,
      preset,
      extraExploreState,
      legacyNotSupported,
    } of TestCases) {
      if (legacyNotSupported) continue;
      it(title, async () => {
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
        metricsExplorerStore.init(
          AD_BIDS_EXPLORE_NAME,
          getInitExploreStateForTest(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            AD_BIDS_TIME_RANGE_SUMMARY,
          ),
        );
        const defaultExplorePreset = getDefaultExplorePreset(
          explore,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        );

        const initState = getCleanMetricsExploreForAssertion();
        await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);
        const curState = getCleanMetricsExploreForAssertion() as ExploreState;

        const url = new URL("http://localhost");
        // load url with legacy protobuf state
        url.searchParams.set(
          "state",
          getProtoFromDashboardState(curState, explore),
        );
        // get back the entity from url params
        const { partialExploreState: entityFromUrl } =
          convertURLSearchParamsToExploreState(
            url.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            defaultExplorePreset,
          );
        expect(entityFromUrl).toEqual({
          ...curState,
          ...(extraExploreState ?? {}),
        });

        // go back to default url
        const defaultUrl = new URL("http://localhost");
        const { partialExploreState: entityFromDefaultUrl } =
          convertURLSearchParamsToExploreState(
            defaultUrl.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            defaultExplorePreset,
          );

        // assert that the entity we got back matches the original
        expect(entityFromDefaultUrl).toEqual(initState);
      });
    }
  });

  it("Large state gets compressed", async () => {
    metricsExplorerStore.init(
      AD_BIDS_EXPLORE_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      ),
    );
    const defaultExploreUrlSearch = getRillDefaultExploreUrlParams(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );
    const defaultExplorePreset = getDefaultExplorePreset(
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_METRICS_INIT,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );

    await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, [
      AD_BIDS_APPLY_LARGE_FILTERS,
      AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
      AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
      AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
    ]);

    // load url params with updated metrics state
    const url = new URL("http://localhost");
    url.search = getCleanedUrlParamsForGoto(
      AD_BIDS_EXPLORE_INIT,
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
      getTimeControlState(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
        get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
      ),
      defaultExploreUrlSearch,
      url,
    ).toString();

    // reset the explore state
    applyURLToExploreState(
      new URL("http://localhost"),
      AD_BIDS_EXPLORE_INIT,
      defaultExplorePreset,
    );
    // reapply the compressed url
    applyURLToExploreState(url, AD_BIDS_EXPLORE_INIT, defaultExplorePreset);

    const currentState = getCleanMetricsExploreForAssertion();
    expect(currentState.selectedTimeRange?.name).toEqual(
      TimeRangePreset.LAST_4_WEEKS,
    );
    expect(currentState.selectedComparisonTimeRange?.name).toEqual(
      TimeComparisonOption.CONTIGUOUS,
    );
    expect(currentState.whereFilter).toEqual(AD_BIDS_LARGE_FILTER);
  });
});

export function applyURLToExploreState(
  url: URL,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const { partialExploreState: partialExploreStateDefaultUrl, errors } =
    convertURLSearchParamsToExploreState(
      url.searchParams,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      exploreSpec,
      defaultExplorePreset,
    );
  metricsExplorerStore.mergePartialExplorerEntity(
    AD_BIDS_EXPLORE_NAME,
    partialExploreStateDefaultUrl,
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  );
  return errors;
}

// cleans the metrics explore of any state that is not stored or restored from url state
export function getCleanMetricsExploreForAssertion() {
  // clone the existing state so that any mutations do affect the copy during assertion
  const cleanedState = deepClone(
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
  ) as Partial<ExploreState>;

  delete cleanedState.name;
  delete cleanedState.proto;
  delete cleanedState.dimensionFilterExcludeMode;
  delete cleanedState.temporaryFilterName;
  delete cleanedState.contextColumnWidths;
  if (cleanedState.selectedTimeRange) {
    cleanedState.selectedTimeRange = {
      name: cleanedState.selectedTimeRange?.name ?? ALL_TIME_RANGE_ALIAS,
      interval: cleanedState.selectedTimeRange?.interval,
    } as DashboardTimeControls;
  }
  delete cleanedState.lastDefinedScrubRange;

  // TODO
  delete cleanedState.leaderboardContextColumn;

  return cleanedState;
}
