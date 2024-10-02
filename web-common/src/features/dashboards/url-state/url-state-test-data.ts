import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

export const URLStateTestMetricsExplorerEntity: Partial<MetricsExplorerEntity> =
  {
    visibleMeasureKeys: new Set(["impressions", "bid_price"]),
    allMeasuresVisible: true,
    visibleDimensionKeys: new Set(["publisher", "domain"]),
    allDimensionsVisible: true,

    selectedDimensionName: "",
    selectedTimezone: "UTC",
    sortDirection: 2,

    leaderboardMeasureName: "impressions",

    pivot: {
      active: false,
      activeCell: null,
      columnPage: 1,
      columns: {
        dimension: [],
        measure: [],
      },
      enableComparison: false,
      expanded: {},
      rowJoinType: "nest",
      rowPage: 1,
      rows: {
        dimension: [],
      },
      sorting: [],
    },

    tdd: {
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: "",
      pinIndex: -1,
    },
  };
