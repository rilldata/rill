import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_PRESET,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  AD_BIDS_SET_KATHMANDU_TIMEZONE,
  AD_BIDS_SET_LA_TIMEZONE,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect, beforeAll, beforeEach } from "vitest";

const AllTimeRangeKeys = [
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
    expectedUrl: "http://localhost/?tr=P4W&tz=Asia%2FKathmandu",
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
    expectedUrl: "http://localhost/?tr=P4W&tz=America%2FLos_Angeles",
  },
];

describe("Human readable URL state", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
  });

  describe("Should update url state and restore default state on empty params", () => {
    for (const { title, mutations, preset, expectedUrl } of OverviewTestCases) {
      it(title, () => {
        const url = new URL("http://localhost");
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
      });
    }
  });
});
