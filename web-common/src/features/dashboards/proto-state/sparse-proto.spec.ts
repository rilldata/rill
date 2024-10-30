import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_SCHEMA,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  AD_BIDS_APPLY_IMP_MEASURE_FILTER,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_OPEN_PUB_IMP_PIVOT,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_SET_P7D_TIME_RANGE_FILTER,
  applyMutationsToDashboard,
  type TestDashboardMutation,
  AD_BIDS_APPLY_DOM_DIMENSION_FILTER,
  AD_BIDS_APPLY_BP_MEASURE_FILTER,
  AD_BIDS_SET_P4W_TIME_RANGE_FILTER,
  AD_BIDS_OPEN_DOM_DIMENSION_TABLE,
  AD_BIDS_OPEN_BP_TDD,
  AD_BIDS_OPEN_DOM_BP_PIVOT,
  AD_BIDS_REMOVE_PUB_DIMENSION_FILTER,
  AD_BIDS_REMOVE_IMP_MEASURE_FILTER,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getPartialDashboard,
  resetDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { beforeAll, it, describe, beforeEach, expect } from "vitest";

const TestCases: {
  title: string;
  mutations: TestDashboardMutation[];
  keys: (keyof MetricsExplorerEntity)[];
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
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    resetDashboardStore();
  });

  describe("should reset dashboard store", () => {
    for (const { title, mutations } of TestCases) {
      it(`from ${title}`, () => {
        const dashboard = getDefaultMetricsExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );
        const defaultProto = getProtoFromDashboardState(dashboard);

        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        metricsExplorerStore.syncFromUrl(
          AD_BIDS_EXPLORE_NAME,
          defaultProto,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_SCHEMA,
        );
        assertDashboardEquals(AD_BIDS_EXPLORE_NAME, dashboard);
      });
    }
  });

  describe("should reset partial dashboard store", () => {
    for (const { title, mutations, keys } of TestCases) {
      it(`to ${title}`, () => {
        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);
        const partialDashboard = getPartialDashboard(
          AD_BIDS_EXPLORE_NAME,
          keys,
        );
        const partialProto = getProtoFromDashboardState(partialDashboard);

        applyMutationsToDashboard(
          AD_BIDS_EXPLORE_NAME,
          TestCasesOppositeMutations,
        );

        metricsExplorerStore.syncFromUrl(
          AD_BIDS_EXPLORE_NAME,
          partialProto,
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_SCHEMA,
        );
        assertDashboardEquals(AD_BIDS_EXPLORE_NAME, {
          ...get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
          ...partialDashboard,
        });
      });
    }
  });
});

function assertDashboardEquals(name: string, expected: MetricsExplorerEntity) {
  expect(cleanDashboard(get(metricsExplorerStore).entities[name])).toEqual(
    cleanDashboard(expected),
  );
}

function cleanDashboard(dashboard: MetricsExplorerEntity) {
  const newDash = deepClone(dashboard);
  delete newDash.proto;
  if (newDash.selectedTimeRange) {
    delete (newDash.selectedTimeRange as any).start;
    delete (newDash.selectedTimeRange as any).end;
  }
  return newDash;
}
