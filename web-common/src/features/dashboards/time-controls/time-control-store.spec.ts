import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  AD_BIDS_INIT,
  AD_BIDS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_SOURCE_NAME,
  AD_BIDS_TIMESTAMP_DIMENSION,
  createAdBidsInStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores-test-data";
import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createTimeControlStore,
  TimeControlState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import TimeControlsStoreTest from "@rilldata/web-common/features/dashboards/time-controls/TimeControlsStoreTest.svelte";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { describe, it, expect } from "vitest";
import { render } from "@testing-library/svelte";

describe("time-control-store", () => {
  runtime.set({
    host: "http://localhost",
    instanceId: "default",
  });
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();

  it("Switching from no timestamp column to having one", async () => {
    const { unmount, queryClient, timeControlsStore } =
      initTimeControlStoreTest(AD_BIDS_INIT);
    await new Promise((resolve) => setTimeout(resolve, 10));

    const state = get(timeControlsStore);
    expect(state.isFetching).toBeFalsy();
    expect(state.ready).toBeTruthy();
    assertStartAndEnd(state, undefined, undefined, undefined, undefined);

    dashboardFetchMocks.mockMetricsView(AD_BIDS_NAME, AD_BIDS_INIT_WITH_TIME);
    dashboardFetchMocks.mockTimeRangeSummary(
      AD_BIDS_SOURCE_NAME,
      AD_BIDS_TIMESTAMP_DIMENSION,
      {
        min: "2022-01-01",
        max: "2022-03-31",
      }
    );
    await queryClient.refetchQueries({
      type: "active",
    });
    await new Promise((resolve) => setTimeout(resolve, 10));

    assertStartAndEnd(
      get(timeControlsStore),
      "2022-01-01T00:00:00.000Z",
      "2022-03-31T00:00:00.001Z",
      "2021-12-31T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z"
    );

    dashboardFetchMocks.mockTimeRangeSummary(
      AD_BIDS_SOURCE_NAME,
      AD_BIDS_TIMESTAMP_DIMENSION,
      {
        min: "2023-01-01",
        max: "2023-03-31",
      }
    );
    await queryClient.refetchQueries({
      type: "active",
    });
    await new Promise((resolve) => setTimeout(resolve, 10));
    // Updating time range updates the start and end
    assertStartAndEnd(
      get(timeControlsStore),
      "2023-01-01T00:00:00.000Z",
      "2023-03-31T00:00:00.001Z",
      "2022-12-31T00:00:00.000Z",
      "2023-04-01T00:00:00.000Z"
    );

    unmount();
  });

  it("Switching selected time range", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(
      AD_BIDS_SOURCE_NAME,
      AD_BIDS_TIMESTAMP_DIMENSION,
      {
        min: "2022-01-01",
        max: "2022-03-31",
      }
    );
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_INIT_WITH_TIME
    );
    await new Promise((resolve) => setTimeout(resolve, 10));

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
      name: TimeRangePreset.LAST_24_HOURS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    assertStartAndEnd(
      get(timeControlsStore),
      "2022-03-30T01:00:00.000Z",
      "2022-03-31T01:00:00.000Z",
      "2022-03-30T00:00:00.000Z",
      "2022-03-31T02:00:00.000Z"
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
      name: TimeRangePreset.CUSTOM,
      start: new Date("2022-03-20T01:00:00.000Z"),
      end: new Date("2022-03-22T01:00:00.000Z"),
      interval: V1TimeGrain.TIME_GRAIN_MONTH,
    });
    let state = get(timeControlsStore);
    assertStartAndEnd(
      state,
      "2022-03-20T01:00:00.000Z",
      "2022-03-22T01:00:00.000Z",
      "2022-03-20T00:00:00.000Z",
      "2022-03-22T02:00:00.000Z"
    );
    // invalid time grain of month is reset to hour
    expect(state.selectedTimeRange.interval).toEqual(
      V1TimeGrain.TIME_GRAIN_HOUR
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
      name: TimeRangePreset.LAST_7_DAYS,
      start: new Date("2021-01-01"),
      end: new Date("2021-03-31"),
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    state = get(timeControlsStore);
    // start and end from selected time range is ignored.
    assertStartAndEnd(
      state,
      "2022-03-25T00:00:00.000Z",
      "2022-04-01T00:00:00.000Z",
      "2022-03-24T23:00:00.000Z",
      "2022-04-01T01:00:00.000Z"
    );
    // valid time grain of hour is retained
    expect(state.selectedTimeRange.interval).toEqual(
      V1TimeGrain.TIME_GRAIN_HOUR
    );

    unmount();
  });

  it("Switching selected comparison time range", async () => {
    dashboardFetchMocks.mockTimeRangeSummary(
      AD_BIDS_SOURCE_NAME,
      AD_BIDS_TIMESTAMP_DIMENSION,
      {
        min: "2022-01-01",
        max: "2022-03-31",
      }
    );
    const { unmount, timeControlsStore } = initTimeControlStoreTest(
      AD_BIDS_INIT_WITH_TIME
    );
    await new Promise((resolve) => setTimeout(resolve, 10));

    metricsExplorerStore.displayComparison(AD_BIDS_NAME, true);
    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
      name: TimeRangePreset.LAST_24_HOURS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    });
    metricsExplorerStore.setSelectedComparisonRange(AD_BIDS_NAME, {} as any);
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to default comparison
      "P1D",
      "2022-03-29T01:00:00.000Z",
      "2022-03-30T01:00:00.000Z",
      "2022-03-29T00:00:00.000Z",
      "2022-03-30T02:00:00.000Z"
    );

    metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
      name: TimeRangePreset.LAST_12_MONTHS,
      start: undefined,
      end: undefined,
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    });
    metricsExplorerStore.setSelectedComparisonRange(AD_BIDS_NAME, {
      name: "P12M",
    } as any);
    assertComparisonStartAndEnd(
      get(timeControlsStore),
      // Sets to the one selected
      "P12M",
      "2020-04-01T00:00:00.000Z",
      "2021-04-01T00:00:00.000Z",
      "2020-03-23T00:00:00.000Z",
      "2021-04-05T00:00:00.000Z"
    );

    unmount();
  });

  function initTimeControlStoreTest(resp: V1MetricsView) {
    createAdBidsInStore();
    dashboardFetchMocks.mockMetricsView(AD_BIDS_NAME, resp);

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          refetchOnMount: false,
          refetchOnReconnect: false,
          refetchOnWindowFocus: false,
          retry: false,
        },
      },
    });
    const stateManagers = createStateManagers({
      queryClient,
      metricsViewName: AD_BIDS_NAME,
    });
    const timeControlsStore = createTimeControlStore(stateManagers);

    const { unmount } = render(TimeControlsStoreTest, {
      timeControlsStore,
    });

    return { unmount, queryClient, timeControlsStore };
  }
});

function assertStartAndEnd(
  timeControlsSate: TimeControlState,
  start: string,
  end: string,
  adjustedStart: string,
  adjustedEnd: string
) {
  expect(timeControlsSate.timeStart).toEqual(start);
  expect(timeControlsSate.timeEnd).toEqual(end);
  expect(timeControlsSate.adjustedStart).toEqual(adjustedStart);
  expect(timeControlsSate.adjustedEnd).toEqual(adjustedEnd);
}

function assertComparisonStartAndEnd(
  timeControlsSate: TimeControlState,
  name: string,
  start: string,
  end: string,
  adjustedStart: string,
  adjustedEnd: string
) {
  expect(timeControlsSate.selectedComparisonTimeRange?.name).toBe(name);
  expect(timeControlsSate.comparisonTimeStart).toBe(start);
  expect(timeControlsSate.comparisonTimeEnd).toBe(end);
  expect(timeControlsSate.comparisonAdjustedStart).toBe(adjustedStart);
  expect(timeControlsSate.comparisonAdjustedEnd).toBe(adjustedEnd);
}
