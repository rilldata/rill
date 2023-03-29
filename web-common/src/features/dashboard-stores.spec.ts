import { describe, it } from "@jest/globals";
import {
  AdBidsMirrorName,
  AdBidsDomainDimension,
  AdBidsName,
  AdBidsPublisherDimension,
  assertMetricsView,
  createAdBidsMirrorInStore,
  createAdBidsInStore,
  TestTimeConstants,
  AdBidsBidPriceMeasure,
} from "@rilldata/web-common/features/dashboard-stores-test-data";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { CUSTOM } from "@rilldata/web-common/lib/time/config";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

const BaseFilter = {
  include: [
    {
      name: AdBidsPublisherDimension,
      in: ["Google", "Facebook"],
    },
    {
      name: AdBidsDomainDimension,
      in: ["google.com"],
    },
  ],
  exclude: [],
};
const ExcludedFilter = {
  include: [
    {
      name: AdBidsDomainDimension,
      in: ["google.com"],
    },
  ],
  exclude: [
    {
      name: AdBidsPublisherDimension,
      in: ["Google", "Facebook"],
    },
  ],
};
const ClearedFilter = {
  include: [],
  exclude: [
    {
      name: AdBidsPublisherDimension,
      in: ["Google", "Facebook"],
    },
  ],
};

// parsed time controls won't have start & end
const AllTimeParsedControls = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
} as DashboardTimeControls;

const Last6HoursControls = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: TestTimeConstants.Last6Hours,
  end: TestTimeConstants.Now,
} as DashboardTimeControls;
// parsed time controls won't have start & end
const Last6HoursParsedControls = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
} as DashboardTimeControls;

const CustomControls = {
  name: TimeRangePreset.CUSTOM,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
  start: TestTimeConstants.Last18Hours,
  end: TestTimeConstants.Last12Hours,
} as DashboardTimeControls;

describe("dashboard-stores", () => {
  it("Toggle filters", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // add filters
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsPublisherDimension,
      "Google"
    );
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsPublisherDimension,
      "Facebook"
    );
    metricsExplorerStore.toggleFilter(
      AdBidsName,
      AdBidsDomainDimension,
      "google.com"
    );
    assertMetricsView(AdBidsName, BaseFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(AdBidsMirrorName, BaseFilter, AllTimeParsedControls);

    // toggle to exclude
    metricsExplorerStore.toggleFilterMode(AdBidsName, AdBidsPublisherDimension);
    assertMetricsView(AdBidsName, ExcludedFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(AdBidsMirrorName, ExcludedFilter, AllTimeParsedControls);

    // clear for Domain
    metricsExplorerStore.clearFilterForDimension(
      AdBidsName,
      AdBidsDomainDimension,
      true
    );
    assertMetricsView(AdBidsName, ClearedFilter);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(AdBidsMirrorName, ClearedFilter, AllTimeParsedControls);

    // clear
    metricsExplorerStore.clearFilters(AdBidsName);
    assertMetricsView(AdBidsName);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore();
    assertMetricsView(AdBidsMirrorName, undefined, AllTimeParsedControls);
  });

  it("Update time selections", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // select a different time
    metricsExplorerStore.setSelectedTimeRange(AdBidsName, Last6HoursControls);
    assertMetricsView(AdBidsName, undefined, Last6HoursControls);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are not persisted
    assertMetricsView(AdBidsMirrorName, undefined, Last6HoursParsedControls);

    // select custom time
    metricsExplorerStore.setSelectedTimeRange(AdBidsName, CustomControls);
    assertMetricsView(AdBidsName, undefined, CustomControls);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore();
    // start and end are persisted for custom
    assertMetricsView(AdBidsMirrorName, undefined, CustomControls);
  });

  it("Select different measure", () => {
    createAdBidsInStore();
    assertMetricsView(AdBidsName);

    // select a different leaderboard measure
    metricsExplorerStore.setLeaderboardMeasureName(
      AdBidsName,
      AdBidsBidPriceMeasure
    );
    assertMetricsView(AdBidsName, undefined, undefined, AdBidsBidPriceMeasure);

    // create a mirror using the proto and assert that the leaderboard measure is persisted
    createAdBidsMirrorInStore();
    assertMetricsView(
      AdBidsMirrorName,
      undefined,
      AllTimeParsedControls,
      AdBidsBidPriceMeasure
    );
  });
});
