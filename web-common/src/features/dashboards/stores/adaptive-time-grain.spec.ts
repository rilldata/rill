import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT_WITH_TIME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { initAdBidsInStore } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  TimeRangePreset,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";

function selectedTimeRange() {
  return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME]
    .selectedTimeRange;
}

function isAdaptive() {
  return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME]
    .adaptiveTimeGrain;
}

const P7D: TimeRange = {
  name: "P7D",
  start: new Date("2022-03-24T00:00:00Z"),
  end: new Date("2022-03-31T00:00:00Z"),
};
const THREE_DAYS_CUSTOM: TimeRange = {
  name: TimeRangePreset.CUSTOM,
  start: new Date("2022-03-28T00:00:00Z"),
  end: new Date("2022-03-31T00:00:00Z"),
};
const NINETY_DAYS_CUSTOM: TimeRange = {
  name: TimeRangePreset.CUSTOM,
  start: new Date("2022-01-01T00:00:00Z"),
  end: new Date("2022-03-31T00:00:00Z"),
};

describe("adaptive time grain in the explore store", () => {
  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    initAdBidsInStore();
  });

  it("is adaptive by default and follows the selected range", () => {
    expect(isAdaptive()).toBe(true);

    // The caller's grain is only a fallback: the rill-time precision wins.
    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      P7D,
      V1TimeGrain.TIME_GRAIN_HOUR,
      undefined,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_DAY);

    // Custom ranges are sized to their duration.
    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      THREE_DAYS_CUSTOM,
      V1TimeGrain.TIME_GRAIN_DAY,
      undefined,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_HOUR);

    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      NINETY_DAYS_CUSTOM,
      V1TimeGrain.TIME_GRAIN_HOUR,
      undefined,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_WEEK);
  });

  it("keeps a fixed grain across range changes", () => {
    metricsExplorerStore.setTimeGrain(
      AD_BIDS_EXPLORE_NAME,
      V1TimeGrain.TIME_GRAIN_HOUR,
    );
    expect(isAdaptive()).toBe(false);
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_HOUR);

    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      P7D,
      V1TimeGrain.TIME_GRAIN_HOUR,
      undefined,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(isAdaptive()).toBe(false);
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
  });

  it("re-derives the grain when switching back to adaptive", () => {
    metricsExplorerStore.setTimeGrain(
      AD_BIDS_EXPLORE_NAME,
      V1TimeGrain.TIME_GRAIN_HOUR,
    );
    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      P7D,
      V1TimeGrain.TIME_GRAIN_HOUR,
      undefined,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_HOUR);

    metricsExplorerStore.setAdaptiveTimeGrain(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
    );
    expect(isAdaptive()).toBe(true);
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_DAY);
  });

  it("respects the smallest grain of the metrics view", () => {
    metricsExplorerStore.selectTimeRange(
      AD_BIDS_EXPLORE_NAME,
      THREE_DAYS_CUSTOM,
      V1TimeGrain.TIME_GRAIN_HOUR,
      undefined,
      {
        ...AD_BIDS_METRICS_INIT_WITH_TIME,
        smallestTimeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      },
    );
    expect(selectedTimeRange()?.interval).toBe(V1TimeGrain.TIME_GRAIN_DAY);
  });
});
