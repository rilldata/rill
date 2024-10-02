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
import { getMetricsExplorerFromUrl } from "@rilldata/web-common/features/dashboards/url-state/fromUrl";
import { getUrlFromMetricsExplorer } from "@rilldata/web-common/features/dashboards/url-state/toUrl";
import { URLStateTestMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/url-state/url-state-test-data";
import {
  DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const NoPresetTestCases: {
  title: string;
  url: string;
  entity: Partial<MetricsExplorerEntity>;
}[] = [
  {
    title: "filter",
    url: "http://localhost/?f=%28publisher+IN+%28%27Yahoo%27%29%29",
    entity: {
      whereFilter: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
      ]),
      dimensionThresholdFilters: [],
    },
  },
  {
    title: "time ranges",
    url: "http://localhost/?tr=P4W&tz=Asia%2FKathmandu",
    entity: {
      selectedTimeRange: {
        name: TimeRangePreset.LAST_4_WEEKS,
      } as DashboardTimeControls,
      selectedTimezone: "Asia/Kathmandu",
    },
  },
  {
    title: "partially visible measures/dimensions",
    url: "http://localhost/?o.m=impressions&o.d=publisher",
    entity: {
      visibleMeasureKeys: new Set([AD_BIDS_IMPRESSIONS_MEASURE]),
      allMeasuresVisible: false,
      visibleDimensionKeys: new Set([AD_BIDS_PUBLISHER_DIMENSION]),
      allDimensionsVisible: false,
    },
  },
  {
    title: "all measures/dimensions visible",
    url: "http://localhost/",
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
  },
  {
    title: "leaderboard sort measure different than default",
    url: "http://localhost/?o.sb=bid_price&o.sd=ASC",
    entity: {
      leaderboardMeasureName: AD_BIDS_BID_PRICE_MEASURE,
      sortDirection: SortDirection.ASCENDING,
    },
  },
  {
    title: "leaderboard sort measure same as default",
    url: "http://localhost/",
    entity: {
      leaderboardMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: SortDirection.DESCENDING,
    },
  },
  {
    title: "time dimensional details",
    url: "http://localhost/?vw=time_dimension&tdd.m=impressions&tdd.ct=stacked_bar",
    entity: {
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      tdd: {
        expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        chartType: TDDChart.STACKED_BAR,
        pinIndex: -1,
      },
    },
  },
  {
    title: "pivot",
    url: "http://localhost/?vw=pivot&p.r=publisher%2Ctime.hour&p.c=domain%2Ctime.day%2Cimpressions",
    entity: {
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
    },
  },
];

describe("Human readable URL state", () => {
  describe("No preset", () => {
    for (const { title, url, entity } of NoPresetTestCases) {
      it(title, () => {
        const u = new URL("http://localhost");
        const defaultEntity = {
          ...getDefaultMetricsExplorerEntity(
            AD_BIDS_NAME,
            AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
            AD_BIDS_EXPLORE_INIT,
            undefined,
          ),
          ...entity,
        };
        getUrlFromMetricsExplorer(
          defaultEntity,
          u.searchParams,
          AD_BIDS_EXPLORE_INIT,
          {},
        );

        expect(u.toString()).toEqual(url);

        const { entity: actualEntity } = getMetricsExplorerFromUrl(
          u.searchParams,
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          AD_BIDS_EXPLORE_INIT,
          {},
        );
        expect(actualEntity).toEqual({
          ...URLStateTestMetricsExplorerEntity,
          ...entity,
        });
      });
    }
  });
});
