import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
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
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { convertPresetToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { convertURLToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import { getUrlFromMetricsExplorer } from "@rilldata/web-common/features/dashboards/url-state/toUrl";
import { URLStateTestMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/url-state/url-state-test-data";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("Human readable URL state", () => {
  it("filter", () => {
    testEntity(
      {
        whereFilter: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
        ]),
        dimensionThresholdFilters: [],
      },
      "http://localhost/?f=%28publisher+IN+%28%27Yahoo%27%29%29",
    );
  });

  describe("Time ranges", () => {
    it("no preset", () => {
      testEntity(
        {
          selectedTimeRange: {
            name: TimeRangePreset.LAST_4_WEEKS,
          } as DashboardTimeControls,
          selectedTimezone: "Asia/Kathmandu",
        },
        "http://localhost/?tr=P4W&tz=Asia%2FKathmandu",
      );
    });

    it("with preset and matching preset", () => {
      testEntity(
        {
          selectedTimeRange: {
            name: TimeRangePreset.LAST_7_DAYS,
          } as DashboardTimeControls,
          selectedTimezone: "Asia/Kathmandu",
        },
        "http://localhost/",
        {
          timeRange: "P7D",
          timezone: "Asia/Kathmandu",
        },
      );
    });

    it("with preset and not matching preset", () => {
      testEntity(
        {
          selectedTimeRange: {
            name: TimeRangePreset.LAST_4_WEEKS,
          } as DashboardTimeControls,
          selectedTimezone: "America/Los_Angeles",
        },
        "http://localhost/?tr=P4W&tz=America%2FLos_Angeles",
        {
          timeRange: "P7D",
          timezone: "Asia/Kathmandu",
        },
      );
    });
  });

  describe("measures/dimensions visibility", () => {
    it("no preset and partially visible measures/dimensions", () => {
      testEntity(
        {
          visibleMeasureKeys: new Set([AD_BIDS_IMPRESSIONS_MEASURE]),
          allMeasuresVisible: false,
          visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
          allDimensionsVisible: false,
        },
        "http://localhost/?o.m=impressions&o.d=publisher",
      );
    });

    it("no preset and all measures/dimensions visible", () => {
      testEntity(
        {
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
        "http://localhost/",
      );
    });

    it("with preset and partially visible measures/dimensions, matching preset", () => {
      testEntity(
        {
          visibleMeasureKeys: new Set([AD_BIDS_IMPRESSIONS_MEASURE]),
          allMeasuresVisible: false,
          visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
          allDimensionsVisible: false,
        },
        "http://localhost/",
        {
          measures: [AD_BIDS_IMPRESSIONS_MEASURE],
          dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        },
      );
    });

    it("with preset and all measures/dimensions visible, not matching preset", () => {
      testEntity(
        {
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
        "http://localhost/?o.m=*&o.d=*",
        {
          measures: [AD_BIDS_IMPRESSIONS_MEASURE],
          dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        },
      );
    });
  });

  describe("leaderboard configs", () => {
    it("no preset and leaderboard sort measure different than default", () => {
      testEntity(
        {
          leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
          sortDirection: SortDirection.ASCENDING,
        },
        "http://localhost/?o.sb=bid_price&o.sd=ASC",
      );
    });

    it("no preset and leaderboard sort measure same as default", () => {
      testEntity(
        {
          leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
          sortDirection: SortDirection.DESCENDING,
        },
        "http://localhost/",
      );
    });

    it("with preset and leaderboard sort measure same as preset", () => {
      testEntity(
        {
          leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
          sortDirection: SortDirection.ASCENDING,
        },
        "http://localhost/",
        {
          overviewSortBy: AD_BIDS_BID_PRICE_MEASURE,
          overviewSortAsc: true,
        },
      );
    });

    it("with preset and leaderboard sort measure different than preset", () => {
      testEntity(
        {
          leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
          sortDirection: SortDirection.DESCENDING,
        },
        "http://localhost/?o.sb=impressions&o.sd=DESC",
        {
          overviewSortBy: AD_BIDS_BID_PRICE_MEASURE,
          overviewSortAsc: true,
        },
      );
    });
  });

  describe("dimension table", () => {
    it("no preset and with dimension table active", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
        },
        "http://localhost/?o.ed=publisher",
      );
    });

    it("with preset and with dimension table same as preset", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
        },
        "http://localhost/",
        {
          overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
        },
      );
    });

    it("with preset and with dimension table different than preset", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
        },
        "http://localhost/?o.ed=publisher",
        {
          overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
        },
      );
    });

    it("with preset and with no dimension table different than preset", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DEFAULT,
          selectedDimensionName: "",
        },
        "http://localhost/?o.ed=",
        {
          overviewExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
        },
      );
    });
  });

  describe("time dimensional details", () => {
    it("no preset and has time dimensional details", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
          tdd: {
            expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
            chartType: TDDChart.STACKED_BAR,
            pinIndex: -1,
          },
        },
        "http://localhost/?vw=time_dimension&tdd.m=impressions&tdd.ct=stacked_bar",
      );
    });

    it("with preset and has time dimensional details same as presets", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
          tdd: {
            expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
            chartType: TDDChart.STACKED_BAR,
            pinIndex: -1,
          },
        },
        "http://localhost/",
        {
          view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_TIME_DIMENSION,
          timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
          timeDimensionChartType: "stacked_bar",
        },
      );
    });

    it("with preset and has time dimensional details different than presets", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DEFAULT,
          tdd: {
            expandedMeasureName: "",
            chartType: TDDChart.DEFAULT,
            pinIndex: -1,
          },
        },
        "http://localhost/?vw=overview&tdd.m=&tdd.ct=timeseries",
        {
          view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_TIME_DIMENSION,
          timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
          timeDimensionChartType: "stacked_bar",
        },
      );
    });
  });

  describe("pivot", () => {
    const PIVOT_ENTITY: Partial<MetricsExplorerEntity> = {
      activePage: DashboardState_ActivePage.PIVOT,
      pivot: {
        active: true,
        rows: {
          dimension: [
            {
              id: AD_BIDS_PUBLISHER_DIMENSION,
              type: PivotChipType.Dimension,
              title: AD_BIDS_PUBLISHER_DIMENSION,
            },
            {
              id: V1TimeGrain.TIME_GRAIN_HOUR,
              type: PivotChipType.Time,
              title: "hour",
            },
          ],
        },
        columns: {
          measure: [
            {
              id: AD_BIDS_IMPRESSIONS_MEASURE,
              type: PivotChipType.Measure,
              title: AD_BIDS_IMPRESSIONS_MEASURE,
            },
          ],
          dimension: [
            {
              id: AD_BIDS_DOMAIN_DIMENSION,
              type: PivotChipType.Dimension,
              title: AD_BIDS_DOMAIN_DIMENSION,
            },
            {
              id: V1TimeGrain.TIME_GRAIN_DAY,
              type: PivotChipType.Time,
              title: "day",
            },
          ],
        },
        expanded: {},
        sorting: [],
        columnPage: 1,
        rowPage: 1,
        enableComparison: false,
        activeCell: null,
        rowJoinType: "nest",
      },
    };

    it("no preset with pivot", () => {
      testEntity(
        PIVOT_ENTITY,
        "http://localhost/?vw=pivot&p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
      );
    });

    it("with preset and pivot same as preset", () => {
      testEntity(PIVOT_ENTITY, "http://localhost/", {
        view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_PIVOT,
        pivotRows: ["publisher", "time.hour"],
        pivotCols: ["domain", "time.day", "impressions"],
      });
    });

    it("with preset and pivot different as preset", () => {
      testEntity(
        PIVOT_ENTITY,
        "http://localhost/?p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
        {
          view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_PIVOT,
          pivotRows: ["domain", "time.day"],
          pivotCols: ["impressions"],
        },
      );
    });

    it("with preset and no pivot, different as preset", () => {
      testEntity(
        {
          activePage: DashboardState_ActivePage.DEFAULT,
          pivot: {
            active: false,
            rows: {
              dimension: [],
            },
            columns: {
              measure: [],
              dimension: [],
            },
            expanded: {},
            sorting: [],
            columnPage: 1,
            rowPage: 1,
            enableComparison: false,
            activeCell: null,
            rowJoinType: "nest",
          },
        },
        "http://localhost/?vw=overview&p.r=&p.c=",
        {
          view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_PIVOT,
          pivotRows: ["domain", "time.day"],
          pivotCols: ["impressions"],
        },
      );
    });
  });
});

function testEntity(
  entity: Partial<MetricsExplorerEntity>,
  expectedUrl: string,
  preset?: V1ExplorePreset,
) {
  const url = new URL("http://localhost");
  const explore: V1ExploreSpec = {
    ...AD_BIDS_EXPLORE_INIT,
    ...(preset ? { defaultPreset: preset } : {}),
  };
  const defaultEntity = {
    ...getDefaultMetricsExplorerEntity(
      AD_BIDS_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      explore,
      undefined,
    ),
    ...entity,
  };
  getUrlFromMetricsExplorer(
    defaultEntity,
    url.searchParams,
    explore,
    preset ?? {},
  );

  expect(url.toString()).to.eq(expectedUrl);

  const { preset: presetFromUrl } = convertURLToExplorePreset(
    url.searchParams,
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    explore,
    getBasePreset(explore, {}),
  );
  const { entity: entityFromPreset } = convertPresetToMetricsExplore(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    explore,
    presetFromUrl,
  );

  expect(entityFromPreset).toEqual({
    ...URLStateTestMetricsExplorerEntity,
    ...entity,
  });
}
