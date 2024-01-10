import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_BASE_FILTER,
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_CLEARED_FILTER,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXCLUDE_FILTER,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_INIT,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_WITH_DELETED_DIMENSION,
  ALL_TIME_PARSED_TEST_CONTROLS,
  assertMetricsView,
  assertMetricsViewRaw,
  createAdBidsMirrorInStore,
  createMetricsMetaQueryMock,
  CUSTOM_TEST_CONTROLS,
  initStateManagers,
  LAST_6_HOURS_TEST_CONTROLS,
  LAST_6_HOURS_TEST_PARSED_CONTROLS,
  resetDashboardStore,
  TestTimeConstants,
  TestTimeOffsetConstants,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  MetricsViewSpecComparisonMode,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

describe("dashboard-stores", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_NAME);
    runtime.set({
      instanceId: "",
      jwt: "",
      host: "",
    });
  });

  beforeEach(() => {
    resetDashboardStore();
  });

  it("Toggle filters", () => {
    const mock = createMetricsMetaQueryMock();
    assertMetricsView(AD_BIDS_NAME);
    const { stateManagers } = initStateManagers();
    const {
      actions: {
        filters: { clearAllFilters },
        dimensionsFilter: {
          toggleDimensionValueSelection,
          toggleDimensionFilterMode,
          removeDimensionFilter,
        },
      },
    } = stateManagers;

    // add filters
    toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Google");
    toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Facebook");
    toggleDimensionValueSelection(AD_BIDS_DOMAIN_DIMENSION, "google.com");
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_BASE_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore(get(mock).data);
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_BASE_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS,
    );

    // toggle to exclude
    toggleDimensionFilterMode(AD_BIDS_PUBLISHER_DIMENSION);
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_EXCLUDE_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore(get(mock).data);
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_EXCLUDE_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS,
    );

    // clear for Domain
    removeDimensionFilter(AD_BIDS_DOMAIN_DIMENSION);
    assertMetricsView(AD_BIDS_NAME, AD_BIDS_CLEARED_FILTER);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore(get(mock).data);
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_CLEARED_FILTER,
      ALL_TIME_PARSED_TEST_CONTROLS,
    );

    // clear
    clearAllFilters();
    assertMetricsView(AD_BIDS_NAME);

    // create a mirror using the proto and assert that the filters are persisted
    createAdBidsMirrorInStore(get(mock).data);
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      ALL_TIME_PARSED_TEST_CONTROLS,
    );
  });

  it("Update time selections", () => {
    const mock = createMetricsMetaQueryMock();
    assertMetricsView(AD_BIDS_NAME);

    // select a different time
    metricsExplorerStore.setSelectedTimeRange(
      AD_BIDS_NAME,
      LAST_6_HOURS_TEST_CONTROLS,
    );
    assertMetricsView(AD_BIDS_NAME, undefined, LAST_6_HOURS_TEST_CONTROLS);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore(get(mock).data);
    // start and end are not persisted
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      LAST_6_HOURS_TEST_PARSED_CONTROLS,
    );

    // select custom time
    metricsExplorerStore.setSelectedTimeRange(
      AD_BIDS_NAME,
      CUSTOM_TEST_CONTROLS,
    );
    assertMetricsView(AD_BIDS_NAME, undefined, CUSTOM_TEST_CONTROLS);

    // create a mirror using the proto and assert that the time controls are persisted
    createAdBidsMirrorInStore(get(mock).data);
    // start and end are persisted for custom
    assertMetricsView(AD_BIDS_MIRROR_NAME, undefined, CUSTOM_TEST_CONTROLS);
  });

  it("Select different measure", () => {
    const mock = createMetricsMetaQueryMock();
    const { stateManagers } = initStateManagers();
    const {
      actions: { setLeaderboardMeasureName },
    } = stateManagers;
    assertMetricsView(AD_BIDS_NAME);

    // select a different leaderboard measure
    setLeaderboardMeasureName(AD_BIDS_BID_PRICE_MEASURE);
    assertMetricsView(
      AD_BIDS_NAME,
      undefined,
      undefined,
      AD_BIDS_BID_PRICE_MEASURE,
    );

    // create a mirror using the proto and assert that the leaderboard measure is persisted
    createAdBidsMirrorInStore(get(mock).data);
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      undefined,
      ALL_TIME_PARSED_TEST_CONTROLS,
      AD_BIDS_BID_PRICE_MEASURE,
    );
  });

  it("Should work when time range is not available", () => {
    const AD_BIDS_NO_TIMESTAMP_NAME = "AdBids_no_timestamp";
    const { stateManagers } = initStateManagers();
    const {
      actions: {
        dimensionsFilter: { toggleDimensionValueSelection },
      },
    } = stateManagers;
    stateManagers.setMetricsViewName(AD_BIDS_NO_TIMESTAMP_NAME);
    metricsExplorerStore.init(
      AD_BIDS_NO_TIMESTAMP_NAME,
      AD_BIDS_INIT,
      undefined,
    );
    assertMetricsViewRaw(
      AD_BIDS_NO_TIMESTAMP_NAME,
      createAndExpression([]),
      undefined,
      AD_BIDS_IMPRESSIONS_MEASURE,
    );

    // add filters
    toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Google");
    toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Facebook");
    toggleDimensionValueSelection(AD_BIDS_DOMAIN_DIMENSION, "google.com");
    assertMetricsViewRaw(
      AD_BIDS_NO_TIMESTAMP_NAME,
      AD_BIDS_BASE_FILTER,
      undefined,
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
  });

  it("Should set the selected time range from the default in config", () => {
    metricsExplorerStore.remove(AD_BIDS_NAME);
    metricsExplorerStore.init(
      AD_BIDS_NAME,
      {
        ...AD_BIDS_INIT,
        defaultTimeRange: "PT6H",
        defaultComparisonMode:
          MetricsViewSpecComparisonMode.COMPARISON_MODE_UNSPECIFIED,
      },
      {
        timeRangeSummary: {
          min: TestTimeConstants.LAST_DAY.toISOString(),
          max: TestTimeConstants.NOW.toISOString(),
          interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
        },
      },
    );

    let metrics = get(metricsExplorerStore).entities[AD_BIDS_NAME];
    // unspecified mode will default to time comparison
    expect(metrics.showTimeComparison).toBeTruthy();
    expect(metrics.selectedComparisonTimeRange?.name).toBe("rill-PP");
    expect(metrics.selectedComparisonTimeRange.start).toEqual(
      TestTimeOffsetConstants.LAST_12_HOURS,
    );
    expect(metrics.selectedComparisonTimeRange.end).toEqual(
      TestTimeOffsetConstants.LAST_6_HOURS,
    );
    expect(metrics.selectedComparisonDimension).toBeUndefined();

    metricsExplorerStore.remove(AD_BIDS_NAME);
    metricsExplorerStore.init(
      AD_BIDS_NAME,
      {
        ...AD_BIDS_INIT,
        defaultTimeRange: "PT6H",
        defaultComparisonMode:
          MetricsViewSpecComparisonMode.COMPARISON_MODE_DIMENSION,
      },
      {
        timeRangeSummary: {
          min: TestTimeConstants.LAST_DAY.toISOString(),
          max: TestTimeConstants.NOW.toISOString(),
          interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
        },
      },
    );
    metrics = get(metricsExplorerStore).entities[AD_BIDS_NAME];
    expect(metrics.showTimeComparison).toBeFalsy();
    // defaults to 1st dimension
    expect(metrics.selectedComparisonDimension).toBe(
      AD_BIDS_PUBLISHER_DIMENSION,
    );

    metricsExplorerStore.remove(AD_BIDS_NAME);
    metricsExplorerStore.init(
      AD_BIDS_NAME,
      {
        ...AD_BIDS_INIT,
        defaultTimeRange: "PT6H",
        defaultComparisonMode:
          MetricsViewSpecComparisonMode.COMPARISON_MODE_DIMENSION,
        defaultComparisonDimension: AD_BIDS_DOMAIN_DIMENSION,
      },
      {
        timeRangeSummary: {
          min: TestTimeConstants.LAST_DAY.toISOString(),
          max: TestTimeConstants.NOW.toISOString(),
          interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
        },
      },
    );
    metrics = get(metricsExplorerStore).entities[AD_BIDS_NAME];
    expect(metrics.selectedComparisonDimension).toBe(AD_BIDS_DOMAIN_DIMENSION);
  });

  describe("Restore invalid state", () => {
    it("Restore invalid filter", () => {
      const mock = createMetricsMetaQueryMock();
      const { stateManagers } = initStateManagers();
      const {
        actions: {
          dimensionsFilter: { toggleDimensionValueSelection },
        },
      } = stateManagers;
      toggleDimensionValueSelection(AD_BIDS_PUBLISHER_DIMENSION, "Facebook");
      toggleDimensionValueSelection(AD_BIDS_DOMAIN_DIMENSION, "google.com");

      // create a mirror from state
      createAdBidsMirrorInStore(get(mock).data);
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(
        AD_BIDS_MIRROR_NAME,
        AD_BIDS_WITH_DELETED_DIMENSION,
      );
      // assert that the filter for removed dimension is not present anymore
      assertMetricsView(
        AD_BIDS_MIRROR_NAME,
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook"]),
        ]),
        ALL_TIME_PARSED_TEST_CONTROLS,
      );
    });

    it("Restore invalid leaderboard measure", () => {
      const mock = createMetricsMetaQueryMock();
      const { stateManagers } = initStateManagers();
      const {
        actions: { setLeaderboardMeasureName },
      } = stateManagers;
      setLeaderboardMeasureName(AD_BIDS_BID_PRICE_MEASURE);

      // create a mirror from state
      createAdBidsMirrorInStore(get(mock).data);
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(AD_BIDS_MIRROR_NAME, {
        name: "AdBids",
        measures: [
          {
            name: AD_BIDS_IMPRESSIONS_MEASURE,
            expression: "count(*)",
          },
        ],
        dimensions: [
          {
            name: AD_BIDS_PUBLISHER_DIMENSION,
          },
        ],
      });
      // assert that the selected measure is reset to the 1st available one
      expect(
        get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .leaderboardMeasureName,
      ).toBe(AD_BIDS_IMPRESSIONS_MEASURE);
    });

    it("Restore invalid selected dimension", () => {
      const mock = createMetricsMetaQueryMock();
      metricsExplorerStore.setMetricDimensionName(
        AD_BIDS_NAME,
        AD_BIDS_DOMAIN_DIMENSION,
      );

      // create a mirror from state
      createAdBidsMirrorInStore(get(mock).data);
      // update the mirrored dashboard mimicking meta query update
      metricsExplorerStore.sync(
        AD_BIDS_MIRROR_NAME,
        AD_BIDS_WITH_DELETED_DIMENSION,
      );
      // assert that the selected dimension is cleared
      expect(
        get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME]
          .selectedDimensionName,
      ).toBeUndefined();
    });
  });
});
