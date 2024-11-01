import { QueryClient } from "@rilldata/svelte-query";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_DEFAULT_TIME_RANGE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_SCHEMA,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { deepClone } from "@vitest/utils";
import { get } from "svelte/store";
import { expect } from "vitest";

export function resetDashboardStore() {
  metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
  initAdBidsInStore();
  initAdBidsMirrorInStore();
}

export function initAdBidsInStore() {
  metricsExplorerStore.init(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_METRICS_INIT,
    AD_BIDS_EXPLORE_INIT,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
}

export function initAdBidsMirrorInStore() {
  metricsExplorerStore.init(
    AD_BIDS_MIRROR_NAME,
    AD_BIDS_METRICS_INIT,
    AD_BIDS_EXPLORE_INIT,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
}

export function createDashboardState(
  name: string,
  metrics: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  whereFilter: V1Expression = createAndExpression([]),
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE,
): MetricsExplorerEntity {
  const explorer = getDefaultMetricsExplorerEntity(
    name,
    metrics,
    explore,
    undefined,
  );
  explorer.whereFilter = whereFilter;
  explorer.selectedTimeRange = timeRange;
  return explorer;
}

export function createAdBidsMirrorInStore({
  metricsView,
  explore,
}: ExploreValidSpecResponse) {
  const proto =
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].proto ?? "";
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AD_BIDS_MIRROR_NAME,
    proto,
    metricsView ?? { measures: [], dimensions: [] },
    explore ?? { metricsView: AD_BIDS_NAME, measures: [], dimensions: [] },
    AD_BIDS_SCHEMA,
  );
}

// Wrapper function to simplify assert call
export function assertMetricsView(
  name: string,
  filters: V1Expression = createAndExpression([]),
  timeRange: DashboardTimeControls = AD_BIDS_DEFAULT_TIME_RANGE,
  selectedMeasure = AD_BIDS_IMPRESSIONS_MEASURE,
) {
  assertMetricsViewRaw(name, filters, timeRange, selectedMeasure);
}

// Raw assert function without any optional params.
// This allows us to assert for `undefined`
// TODO: find a better solution that this hack
export function assertMetricsViewRaw(
  name: string,
  filters: V1Expression,
  timeRange: DashboardTimeControls | undefined,
  selectedMeasure: string,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.whereFilter).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}

export function assertVisiblePartsOfMetricsView(
  name: string,
  measures: Array<string> | undefined,
  dimensions: Array<string> | undefined,
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  if (measures)
    expect([...metricsView.visibleMeasureKeys].sort()).toEqual(measures.sort());
  if (dimensions)
    expect([...metricsView.visibleDimensionKeys].sort()).toEqual(
      dimensions.sort(),
    );
}

export function initStateManagers(metricsViewName?: string) {
  metricsViewName ??= AD_BIDS_NAME;
  const exploreName = metricsViewName + "_explore";

  initAdBidsInStore();

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
  const stateManagers = createStateManagers({
    queryClient,
    metricsViewName,
    exploreName,
  });

  return { stateManagers, queryClient };
}

export function getPartialDashboard(
  name: string,
  keys: (keyof MetricsExplorerEntity)[],
) {
  const dashboard = get(metricsExplorerStore).entities[name];
  const partialDashboard = {} as MetricsExplorerEntity;
  keys.forEach(
    (key) => ((partialDashboard as any)[key] = deepClone(dashboard[key])),
  );
  return partialDashboard;
}

export function getPivotedPartialDashboard(
  pivotRowDimensions: string[],
  pivotRowTimeDimensions: V1TimeGrain[],
  pivotColumnMeasures: string[],
  pivotColumnDimensions: string[],
  pivotColumnTimeDimensions: V1TimeGrain[],
): Partial<MetricsExplorerEntity> {
  const hasPivot =
    !!pivotRowDimensions.length ||
    !!pivotRowTimeDimensions.length ||
    !!pivotColumnMeasures.length ||
    !!pivotColumnDimensions.length ||
    !!pivotColumnTimeDimensions.length;
  return {
    activePage: hasPivot
      ? DashboardState_ActivePage.PIVOT
      : DashboardState_ActivePage.DEFAULT,
    pivot: {
      active: hasPivot,
      rows: {
        dimension: [
          ...pivotRowDimensions.map((r) => ({
            id: r,
            type: PivotChipType.Dimension,
            title: r,
          })),
          ...pivotRowTimeDimensions.map((g) => ({
            id: g,
            type: PivotChipType.Time,
            title: TIME_GRAIN[g]?.label.toString(),
          })),
        ],
      },
      columns: {
        measure: pivotColumnMeasures.map((m) => ({
          id: m,
          type: PivotChipType.Measure,
          title: m,
        })),
        dimension: [
          ...pivotColumnDimensions.map((r) => ({
            id: r,
            type: PivotChipType.Dimension,
            title: r,
          })),
          ...pivotColumnTimeDimensions.map((g) => ({
            id: g,
            type: PivotChipType.Time,
            title: TIME_GRAIN[g]?.label.toString(),
          })),
        ],
      },
      expanded: {},
      sorting: [],
      columnPage: 1,
      rowPage: 1,
      enableComparison: true,
      activeCell: null,
      rowJoinType: "nest",
    },
  };
}
