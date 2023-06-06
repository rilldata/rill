import { expect } from "vitest";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1MetricsViewFilter,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export const AD_BIDS_NAME = "AdBids";
export const AD_BIDS_MIRROR_NAME = "AdBids_mirror";

export const AD_BIDS_IMPRESSIONS_MEASURE = "impressions";
export const AD_BIDS_BID_PRICE_MEASURE = "bid_price";
export const AD_BIDS_PUBLISHER_DIMENSION = "publisher";
export const AD_BIDS_DOMAIN_DIMENSION = "domain";

const Hour = 1000 * 60 * 60;
export const TestTimeConstants = {
  NOW: new Date(),
  LAST_6_HOURS: new Date(Date.now() - Hour * 6),
  LAST_12_HOURS: new Date(Date.now() - Hour * 12),
  LAST_18_HOURS: new Date(Date.now() - Hour * 18),
  LAST_DAY: new Date(Date.now() - Hour * 24),
};

export const AD_BIDS_WITH_DELETED_DIMENSION = {
  name: "AdBids",
  measures: [
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      expression: "count(*)",
    },
    {
      name: AD_BIDS_BID_PRICE_MEASURE,
      expression: "sum(bid_price)",
    },
  ],
  dimensions: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
    },
  ],
};

export function clearMetricsExplorerStore() {
  metricsExplorerStore.remove(AD_BIDS_NAME);
  metricsExplorerStore.remove(AD_BIDS_MIRROR_NAME);
}

export function createAdBidsInStore() {
  metricsExplorerStore.sync(AD_BIDS_NAME, {
    name: "AdBids",
    measures: [
      {
        name: AD_BIDS_IMPRESSIONS_MEASURE,
        expression: "count(*)",
      },
      {
        name: AD_BIDS_BID_PRICE_MEASURE,
        expression: "sum(bid_price)",
      },
    ],
    dimensions: [
      {
        name: AD_BIDS_PUBLISHER_DIMENSION,
      },
      {
        name: AD_BIDS_DOMAIN_DIMENSION,
      },
    ],
  });
  // clear everything if already created
  metricsExplorerStore.clearFilters(AD_BIDS_NAME);
  metricsExplorerStore.setSelectedTimeRange(AD_BIDS_NAME, {
    name: TimeRangePreset.ALL_TIME,
    interval: V1TimeGrain.TIME_GRAIN_MINUTE,
    start: TestTimeConstants.LAST_DAY,
    end: TestTimeConstants.NOW,
  });
}

export function createAdBidsMirrorInStore() {
  const proto = get(metricsExplorerStore).entities[AD_BIDS_NAME].proto;
  // actual url is not relevant here
  metricsExplorerStore.syncFromUrl(
    AD_BIDS_MIRROR_NAME,
    new URL(`http://localhost/dashboard?state=${proto}`)
  );
}

export function assertMetricsView(
  name: string,
  filters: V1MetricsViewFilter = {
    include: [],
    exclude: [],
  },
  timeRange: DashboardTimeControls = {
    name: TimeRangePreset.ALL_TIME,
    interval: V1TimeGrain.TIME_GRAIN_MINUTE,
    start: TestTimeConstants.LAST_DAY,
    end: TestTimeConstants.NOW,
  },
  selectedMeasure = AD_BIDS_IMPRESSIONS_MEASURE
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect(metricsView.filters).toEqual(filters);
  expect(metricsView.selectedTimeRange).toEqual(timeRange);
  expect(metricsView.leaderboardMeasureName).toEqual(selectedMeasure);
}

export function assertVisiblePartsOfMetricsView(
  name: string,
  measures: Array<string>,
  dimensions: Array<string>
) {
  const metricsView = get(metricsExplorerStore).entities[name];
  expect([...metricsView.visibleMeasureKeys]).toEqual(measures);
  expect([...metricsView.visibleDimensionKeys]).toEqual(dimensions);
}

export const AD_BIDS_BASE_FILTER = {
  include: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      in: ["google.com"],
    },
  ],
  exclude: [],
};

export const AD_BIDS_EXCLUDE_FILTER = {
  include: [
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      in: ["google.com"],
    },
  ],
  exclude: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
  ],
};

export const AD_BIDS_CLEARED_FILTER = {
  include: [],
  exclude: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      in: ["Google", "Facebook"],
    },
  ],
};

// parsed time controls won't have start & end
export const ALL_TIME_PARSED_TEST_CONTROLS = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
} as DashboardTimeControls;

export const LAST_6_HOURS_TEST_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: TestTimeConstants.LAST_6_HOURS,
  end: TestTimeConstants.NOW,
} as DashboardTimeControls;

// parsed time controls won't have start & end
export const LAST_6_HOURS_TEST_PARSED_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
} as DashboardTimeControls;

export const CUSTOM_TEST_CONTROLS = {
  name: TimeRangePreset.CUSTOM,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
  start: TestTimeConstants.LAST_18_HOURS,
  end: TestTimeConstants.LAST_12_HOURS,
} as DashboardTimeControls;
