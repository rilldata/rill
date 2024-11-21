import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { setLeaderboardMeasureName } from "@rilldata/web-common/features/dashboards/state-managers/actions/core-actions";
import {
  removeDimensionFilter,
  toggleDimensionValueSelection,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters";
import {
  setPrimaryDimension,
  toggleDimensionVisibility,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/dimensions";
import {
  removeMeasureFilter,
  setMeasureFilter,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/measure-filters";
import { toggleMeasureVisibility } from "@rilldata/web-common/features/dashboards/state-managers/actions/measures";
import {
  setSortDescending,
  toggleSort,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/sorting";
import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export type TestDashboardMutation = (mut: DashboardMutables) => void;
export const AD_BIDS_APPLY_PUB_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => toggleDimensionValueSelection(mut, AD_BIDS_PUBLISHER_DIMENSION, "Google");
export const AD_BIDS_REMOVE_PUB_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => removeDimensionFilter(mut, AD_BIDS_PUBLISHER_DIMENSION);
export const AD_BIDS_APPLY_DOM_DIMENSION_FILTER: TestDashboardMutation = (
  mut,
) => toggleDimensionValueSelection(mut, AD_BIDS_DOMAIN_DIMENSION, "google.com");

export const AD_BIDS_APPLY_IMP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  setMeasureFilter(mut, AD_BIDS_PUBLISHER_DIMENSION, {
    measure: AD_BIDS_IMPRESSIONS_MEASURE,
    type: MeasureFilterType.Value,
    operation: MeasureFilterOperation.GreaterThan,
    value1: "10",
    value2: "",
  });
export const AD_BIDS_REMOVE_IMP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  removeMeasureFilter(
    mut,
    AD_BIDS_PUBLISHER_DIMENSION,
    AD_BIDS_IMPRESSIONS_MEASURE,
  );
export const AD_BIDS_APPLY_BP_MEASURE_FILTER: TestDashboardMutation = (mut) =>
  setMeasureFilter(mut, AD_BIDS_DOMAIN_DIMENSION, {
    measure: AD_BIDS_BID_PRICE_MEASURE,
    type: MeasureFilterType.Value,
    operation: MeasureFilterOperation.GreaterThan,
    value1: "10",
    value2: "",
  });

export const AD_BIDS_SET_P7D_TIME_RANGE_FILTER: TestDashboardMutation = () =>
  metricsExplorerStore.selectTimeRange(
    AD_BIDS_NAME,
    { name: TimeRangePreset.LAST_7_DAYS } as any,
    V1TimeGrain.TIME_GRAIN_DAY,
    undefined,
    AD_BIDS_METRICS_INIT,
  );
export const AD_BIDS_SET_P4W_TIME_RANGE_FILTER: TestDashboardMutation = () =>
  metricsExplorerStore.selectTimeRange(
    AD_BIDS_NAME,
    { name: TimeRangePreset.LAST_4_WEEKS } as any,
    V1TimeGrain.TIME_GRAIN_WEEK,
    undefined,
    AD_BIDS_METRICS_INIT,
  );
export const AD_BIDS_SET_KATHMANDU_TIMEZONE: TestDashboardMutation = () =>
  metricsExplorerStore.setTimeZone(AD_BIDS_NAME, "Asia/Kathmandu");
export const AD_BIDS_SET_LA_TIMEZONE: TestDashboardMutation = () =>
  metricsExplorerStore.setTimeZone(AD_BIDS_NAME, "America/Los_Angeles");

export const AD_BIDS_TOGGLE_BID_PRICE_MEASURE: TestDashboardMutation = (
  mut,
) => {
  toggleMeasureVisibility(
    mut,
    AD_BIDS_EXPLORE_INIT.measures!,
    AD_BIDS_BID_PRICE_MEASURE,
  );
};
export const AD_BIDS_TOGGLE_BID_DOMAIN_DIMENSION: TestDashboardMutation = (
  mut,
) => {
  toggleDimensionVisibility(
    mut,
    AD_BIDS_EXPLORE_INIT.dimensions!,
    AD_BIDS_DOMAIN_DIMENSION,
  );
};

export const AD_BIDS_SORT_DESC_BY_IMPRESSIONS: TestDashboardMutation = (
  mut,
) => {
  setLeaderboardMeasureName(mut, AD_BIDS_IMPRESSIONS_MEASURE);
  setSortDescending(mut);
};
export const AD_BIDS_SORT_ASC_BY_BID_PRICE: TestDashboardMutation = (mut) => {
  setLeaderboardMeasureName(mut, AD_BIDS_BID_PRICE_MEASURE);
  setSortDescending(mut);
  toggleSort(mut, mut.dashboard.dashboardSortType);
};

export const AD_BIDS_OPEN_PUB_DIMENSION_TABLE: TestDashboardMutation = (mut) =>
  setPrimaryDimension(mut, AD_BIDS_PUBLISHER_DIMENSION);
export const AD_BIDS_OPEN_DOM_DIMENSION_TABLE: TestDashboardMutation = (mut) =>
  setPrimaryDimension(mut, AD_BIDS_DOMAIN_DIMENSION);

export const AD_BIDS_OPEN_IMP_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setExpandedMeasureName(
    AD_BIDS_NAME,
    AD_BIDS_IMPRESSIONS_MEASURE,
  );
export const AD_BIDS_OPEN_BP_TDD: TestDashboardMutation = () =>
  metricsExplorerStore.setExpandedMeasureName(
    AD_BIDS_NAME,
    AD_BIDS_BID_PRICE_MEASURE,
  );

export const AD_BIDS_OPEN_PUB_IMP_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_NAME,
    {
      dimension: [
        {
          id: AD_BIDS_PUBLISHER_DIMENSION,
          title: AD_BIDS_PUBLISHER_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
    },
    {
      dimension: [],
      measure: [
        {
          id: AD_BIDS_IMPRESSIONS_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );
export const AD_BIDS_OPEN_DOM_BP_PIVOT: TestDashboardMutation = () =>
  metricsExplorerStore.createPivot(
    AD_BIDS_NAME,
    {
      dimension: [
        {
          id: AD_BIDS_DOMAIN_DIMENSION,
          title: AD_BIDS_DOMAIN_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
    },
    {
      dimension: [],
      measure: [
        {
          id: AD_BIDS_BID_PRICE_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    },
  );

export function applyMutationsToDashboard(
  name: string,
  mutations: TestDashboardMutation[],
) {
  const dashboard = get(metricsExplorerStore).entities[name];
  const dashboardMutables = { dashboard } as DashboardMutables;

  mutations.forEach((mutation) => mutation(dashboardMutables));
}
