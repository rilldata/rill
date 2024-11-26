import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getPivotedPartialDashboard } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

const TestCases: {
  title: string;
  entity: Partial<MetricsExplorerEntity>;
  preset?: V1ExplorePreset;
  expectedUrl: string;
}[] = [
  {
    title: "filter",
    entity: {
      whereFilter: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
      ]),
      dimensionThresholdFilters: [],
    },
    expectedUrl: "http://localhost/?f=publisher+IN+%28%27Yahoo%27%29",
  },

  {
    title: "Time range without preset",
    entity: {
      selectedTimeRange: {
        name: TimeRangePreset.LAST_4_WEEKS,
      } as DashboardTimeControls,
      selectedTimezone: "Asia/Kathmandu",
    },
    expectedUrl: "http://localhost/?tr=P4W&tz=Asia%2FKathmandu",
  },
  {
    title: "Time range with preset and state matching preset",
    entity: {
      selectedTimeRange: {
        name: TimeRangePreset.LAST_7_DAYS,
      } as DashboardTimeControls,
      selectedTimezone: "Asia/Kathmandu",
    },
    preset: {
      timeRange: "P7D",
      timezone: "Asia/Kathmandu",
    },
    expectedUrl: "http://localhost/",
  },
  {
    title: "Time range with preset and state not matching preset",
    entity: {
      selectedTimeRange: {
        name: TimeRangePreset.LAST_4_WEEKS,
      } as DashboardTimeControls,
      selectedTimezone: "America/Los_Angeles",
    },
    preset: {
      timeRange: "P7D",
      timezone: "Asia/Kathmandu",
    },
    expectedUrl: "http://localhost/?tr=P4W&tz=America%2FLos_Angeles",
  },

  {
    title:
      "Measures/dimensions visibility with no preset and partially visible measures/dimensions in state",
    entity: {
      visibleMeasureKeys: new Set([AD_BIDS_IMPRESSIONS_MEASURE]),
      allMeasuresVisible: false,
      visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
      allDimensionsVisible: false,
    },
    expectedUrl: "http://localhost/?o.m=impressions&o.d=publisher",
  },
  {
    title:
      "Measures/dimensions visibility with no preset and all measures/dimensions visible in state",
    entity: {
      visibleMeasureKeys: new Set([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]),
      allMeasuresVisible: true,
      visibleDimensionKeys: new Set([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]),
      allDimensionsVisible: true,
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and partially visible measures/dimensions in state matching preset",
    entity: {
      visibleMeasureKeys: new Set([AD_BIDS_IMPRESSIONS_MEASURE]),
      allMeasuresVisible: false,
      visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
      allDimensionsVisible: false,
    },
    preset: {
      measures: [AD_BIDS_IMPRESSIONS_MEASURE],
      dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Measures/dimensions visibility with preset and all measures/dimensions visible in state not matching preset",
    entity: {
      visibleMeasureKeys: new Set([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]),
      allMeasuresVisible: true,
      visibleDimensionKeys: new Set([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]),
      allDimensionsVisible: true,
    },
    preset: {
      measures: [AD_BIDS_IMPRESSIONS_MEASURE],
      dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
    },
    expectedUrl: "http://localhost/?o.m=*&o.d=*",
  },

  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state different than default",
    entity: {
      leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: SortDirection.ASCENDING,
    },
    expectedUrl: "http://localhost/?o.sb=bid_price&o.sd=ASC",
  },
  {
    title:
      "Leaderboard configs with no preset and leaderboard sort measure in state same as default",
    entity: {
      leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: SortDirection.DESCENDING,
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state same as preset",
    entity: {
      leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: SortDirection.ASCENDING,
    },
    preset: {
      overviewSortBy: AD_BIDS_BID_PRICE_MEASURE,
      overviewSortAsc: true,
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Leaderboard configs with preset and leaderboard sort measure in state different than preset",
    entity: {
      leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: SortDirection.DESCENDING,
    },
    preset: {
      overviewSortBy: AD_BIDS_BID_PRICE_MEASURE,
      overviewSortAsc: true,
    },
    expectedUrl: "http://localhost/?o.sb=impressions&o.sd=DESC",
  },

  {
    title: "Dimension table with no preset and dimension table active in state",
    entity: {
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
    },
    expectedUrl: "http://localhost/?o.ed=publisher",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state same as preset",
    entity: {
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
    },
    preset: {
      overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Dimension table with preset and with dimension table in state different than preset",
    entity: {
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
    },
    preset: {
      overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
    },
    expectedUrl: "http://localhost/?o.ed=publisher",
  },
  {
    title:
      "Dimension table with preset and with no dimension table in state different than preset",
    entity: {
      activePage: DashboardState_ActivePage.DEFAULT,
      selectedDimensionName: "",
    },
    preset: {
      overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
    },
    expectedUrl: "http://localhost/?o.ed=",
  },

  {
    title:
      "Time dimensional details with no preset and has time dimensional details in state",
    entity: {
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      tdd: {
        expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        chartType: TDDChart.STACKED_BAR,
        pinIndex: -1,
      },
    },
    expectedUrl:
      "http://localhost/?vw=time_dimension&tdd.m=impressions&tdd.ct=stacked_bar",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state same as presets",
    entity: {
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      tdd: {
        expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        chartType: TDDChart.STACKED_BAR,
        pinIndex: -1,
      },
    },
    preset: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      timeDimensionChartType: "stacked_bar",
    },
    expectedUrl: "http://localhost/",
  },
  {
    title:
      "Time dimensional details with preset and has time dimensional details in state different than presets",
    entity: {
      activePage: DashboardState_ActivePage.DEFAULT,
      tdd: {
        expandedMeasureName: "",
        chartType: TDDChart.DEFAULT,
        pinIndex: -1,
      },
    },
    preset: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
      timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      timeDimensionChartType: "stacked_bar",
    },
    expectedUrl: "http://localhost/?vw=overview&tdd.m=&tdd.ct=timeseries",
  },

  {
    title: "Pivot with no preset and has pivot in state",
    entity: getPivotedPartialDashboard(
      [AD_BIDS_PUBLISHER_DIMENSION],
      [V1TimeGrain.TIME_GRAIN_HOUR],
      [AD_BIDS_IMPRESSIONS_MEASURE],
      [AD_BIDS_DOMAIN_DIMENSION],
      [V1TimeGrain.TIME_GRAIN_DAY],
    ),
    expectedUrl:
      "http://localhost/?vw=pivot&p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
  },
  {
    title: "Pivot with preset and has pivot in state same as preset",
    entity: getPivotedPartialDashboard(
      [AD_BIDS_PUBLISHER_DIMENSION],
      [V1TimeGrain.TIME_GRAIN_HOUR],
      [AD_BIDS_IMPRESSIONS_MEASURE],
      [AD_BIDS_DOMAIN_DIMENSION],
      [V1TimeGrain.TIME_GRAIN_DAY],
    ),
    preset: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      pivotRows: ["publisher", "time.hour"],
      pivotCols: ["domain", "time.day", "impressions"],
    },
    expectedUrl: "http://localhost/",
  },
  {
    title: "Pivot with preset and pivot in state different as preset",
    entity: getPivotedPartialDashboard(
      [AD_BIDS_DOMAIN_DIMENSION],
      [V1TimeGrain.TIME_GRAIN_DAY],
      [AD_BIDS_IMPRESSIONS_MEASURE],
      [],
      [],
    ),
    preset: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      pivotRows: ["publisher", "time.hour"],
      pivotCols: ["domain", "time.day", "impressions"],
    },
    expectedUrl: "http://localhost/?p.r=domain%2Ctime.day&p.c=impressions",
  },
  {
    title: "Pivot with preset and no pivot in state different as preset",
    entity: getPivotedPartialDashboard([], [], [], [], []),
    preset: {
      view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      pivotRows: ["publisher", "time.hour"],
      pivotCols: ["domain", "time.day", "impressions"],
    },
    expectedUrl: "http://localhost/?vw=overview&p.r=&p.c=",
  },
];

describe("Human readable URL state", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
  });

  describe("Should update url state and restore default state on empty params", () => {
    for (const { title, entity, preset, expectedUrl } of TestCases) {
      it(title, () => {
        const url = new URL("http://localhost");
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
        const basePreset = getBasePreset(explore, {
          timeZone: "UTC",
        });
        const initEntity = getDefaultMetricsExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );
        cleanMetricsExplore(initEntity);

        // load url params with update metrics state
        mergeSearchParams(
          convertMetricsEntityToURLSearchParams(
            {
              ...initEntity,
              ...entity,
            },
            explore,
            basePreset,
          ),
          url.searchParams,
        );

        expect(url.toString()).to.eq(expectedUrl);

        // get back the entity from url params
        const { partialExploreState: entityFromUrl } =
          convertURLToMetricsExplore(
            url.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            basePreset,
          );

        // assert that the entity we got back matches the expected entity
        expect(entityFromUrl).toEqual({
          ...initEntity,
          ...entity,
        });

        // go back to default url
        const defaultUrl = new URL("http://localhost");
        const { partialExploreState: entityFromDefaultUrl } =
          convertURLToMetricsExplore(
            defaultUrl.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            basePreset,
          );

        // assert that the entity we got back matches the original
        expect(entityFromDefaultUrl).toEqual(initEntity);
      });
    }
  });

  describe("Should set correct state for legacy protobuf state and restore default state on empty params", () => {
    for (const { title, entity, preset } of TestCases) {
      it(title, () => {
        const url = new URL("http://localhost");
        const explore: V1ExploreSpec = {
          ...AD_BIDS_EXPLORE_INIT,
          ...(preset ? { defaultPreset: preset } : {}),
        };
        const basePreset = getBasePreset(explore, {
          timeZone: "UTC",
        });
        const initEntity = getDefaultMetricsExplorerEntity(
          AD_BIDS_EXPLORE_NAME,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          explore,
          AD_BIDS_TIME_RANGE_SUMMARY,
        );
        cleanMetricsExplore(initEntity);
        // load url with legacy protobuf state
        url.searchParams.set(
          "state",
          getProtoFromDashboardState({
            ...initEntity,
            ...entity,
          }),
        );

        // get back the entity from url params
        const { partialExploreState: entityFromUrl } =
          convertURLToMetricsExplore(
            url.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            basePreset,
          );
        // assert that the entity we got back matches the expected entity
        expect(entityFromUrl).toEqual({
          ...initEntity,
          ...entity,
        });

        // go back to default url
        const defaultUrl = new URL("http://localhost");
        const { partialExploreState: entityFromDefaultUrl } =
          convertURLToMetricsExplore(
            defaultUrl.searchParams,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            explore,
            basePreset,
          );

        // assert that the entity we got back matches the original
        expect(entityFromDefaultUrl).toEqual(initEntity);
      });
    }
  });
});

// cleans up any UI only state from MetricsExplorerEntity
export function cleanMetricsExplore(
  metricsExplorerEntity: Partial<MetricsExplorerEntity>,
) {
  delete metricsExplorerEntity.name;
  delete metricsExplorerEntity.dimensionFilterExcludeMode;
  delete metricsExplorerEntity.temporaryFilterName;
  delete metricsExplorerEntity.contextColumnWidths;
  if (metricsExplorerEntity.selectedTimeRange) {
    metricsExplorerEntity.selectedTimeRange = {
      name: metricsExplorerEntity.selectedTimeRange?.name ?? "inf",
      // TODO: grain
    } as DashboardTimeControls;
  }
  delete metricsExplorerEntity.lastDefinedScrubRange;

  // TODO
  delete metricsExplorerEntity.selectedScrubRange;
  delete metricsExplorerEntity.leaderboardContextColumn;
  delete metricsExplorerEntity.dashboardSortType;
}
