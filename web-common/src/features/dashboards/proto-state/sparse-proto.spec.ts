import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  getInitExploreStateForTest,
  getPartialDashboard,
  resetDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_APPLY_BP_MEASURE_FILTER,
  AD_BIDS_APPLY_DOM_DIMENSION_FILTER,
  AD_BIDS_APPLY_IMP_MEASURE_FILTER,
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_OPEN_BP_TDD,
  AD_BIDS_OPEN_DOM_BP_PIVOT,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_OPEN_PUB_IMP_PIVOT,
  AD_BIDS_REMOVE_IMP_MEASURE_FILTER,
  AD_BIDS_REMOVE_PUB_DIMENSION_FILTER,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";

const TestCases: {
  title: string;
  mutations: TestDashboardMutation[];
  keys: (keyof ExploreState)[];
}[] = [
  {
    title: "filters",
    mutations: [
      AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
      AD_BIDS_APPLY_IMP_MEASURE_FILTER,
    ],
    keys: ["whereFilter", "dimensionThresholdFilters"],
  },
  {
    title: "time range",
    mutations: [AD_BIDS_SET_P7D_TIME_RANGE_FILTER],
    keys: ["selectedTimeRange", "selectedComparisonTimeRange"],
  },
  {
    title: "dimension table",
    mutations: [AD_BIDS_OPEN_PUB_DIMENSION_TABLE],
    keys: ["activePage", "selectedDimensionName"],
  },
  {
    title: "time dimension details",
    mutations: [AD_BIDS_OPEN_IMP_TDD],
    keys: ["activePage", "tdd"],
  },
  {
    title: "pivot",
    mutations: [AD_BIDS_OPEN_PUB_IMP_PIVOT],
    keys: ["activePage", "pivot"],
  },
];
// list of mutations that reverts all mutations from the above test cases
const TestCasesOppositeMutations = [
  AD_BIDS_REMOVE_PUB_DIMENSION_FILTER,
  AD_BIDS_APPLY_DOM_DIMENSION_FILTER,
  AD_BIDS_REMOVE_IMP_MEASURE_FILTER,
  AD_BIDS_APPLY_BP_MEASURE_FILTER,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_BP_TDD,
  AD_BIDS_OPEN_DOM_BP_PIVOT,
];

describe("sparse proto", () => {
  beforeEach(() => {
    resetDashboardStore();
  });

  describe("should reset dashboard store", () => {
    for (const { title, mutations } of TestCases) {
      it(`from ${title}`, async () => {
        const dashboard = getFullInitExploreState(
          AD_BIDS_EXPLORE_NAME,
          getInitExploreStateForTest(
            AD_BIDS_METRICS_INIT,
            AD_BIDS_EXPLORE_INIT,
            AD_BIDS_TIME_RANGE_SUMMARY,
          ),
        );
        const defaultProto = getProtoFromDashboardState(
          dashboard,
          AD_BIDS_EXPLORE_INIT,
        );

        await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        metricsExplorerStore.syncFromUrl(
          AD_BIDS_EXPLORE_NAME,
          defaultProto,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
        );
        assertDashboardEquals(AD_BIDS_EXPLORE_NAME, dashboard);
      });
    }
  });

  describe("should reset partial dashboard store", () => {
    for (const { title, mutations, keys } of TestCases) {
      it(`to ${title}`, async () => {
        await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);
        const partialDashboard = getPartialDashboard(
          AD_BIDS_EXPLORE_NAME,
          keys,
        );
        const partialProto = getProtoFromDashboardState(
          partialDashboard,
          AD_BIDS_EXPLORE_INIT,
        );

        await applyMutationsToDashboard(
          AD_BIDS_EXPLORE_NAME,
          TestCasesOppositeMutations,
        );

        metricsExplorerStore.syncFromUrl(
          AD_BIDS_EXPLORE_NAME,
          partialProto,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
        );
        assertDashboardEquals(AD_BIDS_EXPLORE_NAME, {
          ...get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          ...partialDashboard,
        });
      });
    }
  });
});

function assertDashboardEquals(name: string, expectedState: ExploreState) {
  expect(cleanDashboard(get(metricsExplorerStore).entities[name])).toEqual(
    cleanDashboard(expectedState),
  );
}

function cleanDashboard(exploreState: ExploreState) {
  const newDash = deepClone(exploreState);
  delete newDash.proto;
  if (newDash.selectedTimeRange) {
    delete (newDash.selectedTimeRange as any).start;
    delete (newDash.selectedTimeRange as any).end;
  }
  delete (newDash as any).contextColumnWidths;
  delete (newDash as any).temporaryFilterName;
  return newDash;
}
