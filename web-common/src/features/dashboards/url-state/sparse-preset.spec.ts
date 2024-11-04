import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_BASE_PRESET,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
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
import { convertMetricsExploreToPreset } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsExploreToPreset";
import {
  convertPresetToMetricsExplore,
  convertURLToMetricsExplore,
} from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { cleanMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/url-state.spec";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

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

describe("sparse preset", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    resetDashboardStore();
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
  });

  describe("should reset dashboard store", () => {
    for (const { title, mutations } of TestCases) {
      it(`from ${title}`, () => {
        const initEntity = getDefaultMetricsExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );
        cleanMetricsExplore(initEntity);
        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        const url = new URL("http://localhost");
        // get the entity from no param url
        const { entity: entityFromUrl } = convertURLToMetricsExplore(
          url.searchParams,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_BASE_PRESET,
        );

        expect(entityFromUrl).toEqual(initEntity);
      });
    }
  });

  describe("should reset partial dashboard store", () => {
    for (const { title, mutations, keys } of TestCases) {
      it(`to ${title}`, () => {
        // apply the mutations for the test
        applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

        // get and store the partial from the current state
        const partialEntity = getPartialDashboard(AD_BIDS_EXPLORE_NAME, keys);
        const partialPreset = convertMetricsExploreToPreset(
          partialEntity,
          AD_BIDS_EXPLORE_INIT,
        );
        // convert to partial preset and back to entity
        const { entity: partialEntityFromPreset } =
          convertPresetToMetricsExplore(
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            AD_BIDS_EXPLORE_INIT,
            partialPreset,
          );
        // remove parts not in the state
        cleanMetricsExplore(partialEntity);
        // assert that the partial from entity match the one from state
        expect(partialEntityFromPreset).toEqual(partialEntity);

        // apply mutations to reverse the previous mutations
        applyMutationsToDashboard(
          AD_BIDS_EXPLORE_NAME,
          TestCasesOppositeMutations,
        );

        // merge partial entity from preset to current state
        metricsExplorerStore.mergePartialExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          partialEntityFromPreset,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        );

        // get and store the current state
        const curEntity = deepClone(
          get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME],
        );
        cleanMetricsExplore(curEntity);
        // assert that the current entity is not changed when the partial entity selected is applied to it.
        expect(curEntity).toEqual({
          ...curEntity,
          ...partialEntity,
        });
      });
    }
  });
});
